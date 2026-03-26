package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/urmzd/adk/provider/ollama"
	"github.com/urmzd/kgdk/kgtypes"
	"github.com/urmzd/zoro/internal/models"
	"github.com/urmzd/zoro/internal/searcher"
)

type Orchestrator struct {
	graph       kgtypes.Graph
	adapter     *ollama.Adapter
	searcher    *searcher.Searcher
	sessions    map[string]*models.ResearchSession
	subscribers map[string][]chan models.SSEEvent
	mu          sync.RWMutex
}

func New(g kgtypes.Graph, a *ollama.Adapter, s *searcher.Searcher) *Orchestrator {
	return &Orchestrator{
		graph:       g,
		adapter:     a,
		searcher:    s,
		sessions:    make(map[string]*models.ResearchSession),
		subscribers: make(map[string][]chan models.SSEEvent),
	}
}

func (o *Orchestrator) CreateSession(query string) *models.ResearchSession {
	session := &models.ResearchSession{
		ID:        uuid.New().String(),
		Query:     query,
		Status:    "created",
		Results:   []models.SearchResult{},
		Timeline:  []models.TimelineEvent{},
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	o.sessions[session.ID] = session
	o.mu.Unlock()
	return session
}

func (o *Orchestrator) GetSession(id string) *models.ResearchSession {
	o.mu.RLock()
	defer o.mu.RUnlock()
	s := o.sessions[id]
	if s == nil {
		return nil
	}
	cp := *s
	return &cp
}

func (o *Orchestrator) Subscribe(id string) <-chan models.SSEEvent {
	o.mu.RLock()
	_, exists := o.sessions[id]
	o.mu.RUnlock()
	if !exists {
		return nil
	}

	ch := make(chan models.SSEEvent, 128)
	o.mu.Lock()
	o.subscribers[id] = append(o.subscribers[id], ch)
	o.mu.Unlock()
	return ch
}

func (o *Orchestrator) emit(sessionID string, evt models.SSEEvent) {
	o.mu.Lock()
	if s := o.sessions[sessionID]; s != nil {
		msg, err := json.Marshal(evt.Data)
		if err != nil {
			msg = []byte(fmt.Sprintf("%v", evt.Data))
		}
		s.Timeline = append(s.Timeline, models.TimelineEvent{
			Type:      evt.Type,
			Message:   string(msg),
			Timestamp: time.Now(),
		})
	}
	channels := o.subscribers[sessionID]
	o.mu.Unlock()

	for _, ch := range channels {
		select {
		case ch <- evt:
		default:
			log.Printf("[orchestrator] warning: dropped event %s for session %s (slow subscriber)", evt.Type, sessionID)
		}
	}
}

func (o *Orchestrator) closeSubscribers(sessionID string) {
	o.mu.Lock()
	channels := o.subscribers[sessionID]
	delete(o.subscribers, sessionID)
	o.mu.Unlock()
	for _, ch := range channels {
		close(ch)
	}
}

func (o *Orchestrator) setStatus(sessionID, status string) {
	o.mu.Lock()
	if s := o.sessions[sessionID]; s != nil {
		s.Status = status
	}
	o.mu.Unlock()
}

func (o *Orchestrator) Run(ctx context.Context, sessionID string) {
	o.setStatus(sessionID, "running")

	session := o.GetSession(sessionID)
	if session == nil {
		o.closeSubscribers(sessionID)
		return
	}
	query := session.Query

	// Check context before each major step
	if ctx.Err() != nil {
		o.closeSubscribers(sessionID)
		return
	}

	// Step 1: Prior knowledge
	o.emit(sessionID, models.SSEEvent{
		Type: models.EventPriorKnowledge,
		Data: map[string]string{"message": "Searching prior knowledge..."},
	})

	priorFacts, err := o.graph.SearchFacts(ctx, query)
	if err != nil {
		log.Printf("prior knowledge search error (non-fatal): %v", err)
	} else if len(priorFacts.Facts) > 0 {
		o.emit(sessionID, models.SSEEvent{
			Type: models.EventPriorKnowledge,
			Data: priorFacts.Facts,
		})
	}

	if ctx.Err() != nil {
		o.closeSubscribers(sessionID)
		return
	}

	// Step 2: Web search
	o.emit(sessionID, models.SSEEvent{
		Type: models.EventSearchStarted,
		Data: map[string]string{"query": query},
	})

	results, err := o.searcher.Search(ctx, query)
	if err != nil {
		o.emit(sessionID, models.SSEEvent{
			Type: models.EventError,
			Data: map[string]string{"error": err.Error()},
		})
		o.setStatus(sessionID, "error")
		o.closeSubscribers(sessionID)
		return
	}

	o.mu.Lock()
	if s := o.sessions[sessionID]; s != nil {
		s.Results = results
	}
	o.mu.Unlock()

	o.emit(sessionID, models.SSEEvent{
		Type: models.EventSearchResults,
		Data: results,
	})

	if ctx.Err() != nil {
		o.closeSubscribers(sessionID)
		return
	}

	// Step 3: Ingest each result
	for i, result := range results {
		if ctx.Err() != nil {
			o.closeSubscribers(sessionID)
			return
		}
		episodeBody := fmt.Sprintf("Title: %s\nURL: %s\nSnippet: %s", result.Title, result.URL, result.Snippet)
		input := &kgtypes.EpisodeInput{
			Name:    fmt.Sprintf("%s - Result %d", query, i+1),
			Body:    episodeBody,
			Source:  result.URL,
			GroupID: sessionID,
		}

		resp, err := o.graph.IngestEpisode(ctx, input)
		if err != nil {
			log.Printf("episode ingestion error (result %d): %v", i+1, err)
			continue
		}

		o.emit(sessionID, models.SSEEvent{
			Type: models.EventEpisodeIngested,
			Data: map[string]any{"result_index": i, "episode_uuid": resp.UUID},
		})

		for _, entity := range resp.EntityNodes {
			o.emit(sessionID, models.SSEEvent{
				Type: models.EventEntityDiscovered,
				Data: entity,
			})
		}
		for _, relation := range resp.EpisodicEdges {
			o.emit(sessionID, models.SSEEvent{
				Type: models.EventRelationFound,
				Data: relation,
			})
		}
	}

	if ctx.Err() != nil {
		o.closeSubscribers(sessionID)
		return
	}

	// Step 4: Get session subgraph
	subgraph, err := o.graph.SearchFacts(ctx, query, kgtypes.WithGroupID(sessionID))
	if err != nil {
		log.Printf("subgraph search error: %v", err)
	} else {
		o.emit(sessionID, models.SSEEvent{
			Type: models.EventGraphReady,
			Data: subgraph,
		})

		// Step 5: Stream LLM synthesis
		prompt := buildSummaryPrompt(query, results, subgraph)
		rx, err := o.adapter.GenerateStream(ctx, prompt)
		if err != nil {
			log.Printf("llm stream error: %v", err)
		} else {
			var summary strings.Builder
			for token := range rx {
				if ctx.Err() != nil {
					break
				}
				summary.WriteString(token)
				o.emit(sessionID, models.SSEEvent{
					Type: models.EventSummaryToken,
					Data: map[string]string{"token": token},
				})
			}
			o.mu.Lock()
			if s := o.sessions[sessionID]; s != nil {
				s.Summary = summary.String()
			}
			o.mu.Unlock()
		}
	}

	// Step 6: Complete
	o.setStatus(sessionID, "complete")
	o.emit(sessionID, models.SSEEvent{
		Type: models.EventResearchComplete,
		Data: map[string]string{"session_id": sessionID},
	})
	o.closeSubscribers(sessionID)
}

func buildSummaryPrompt(query string, results []models.SearchResult, facts *kgtypes.SearchFactsResult) string {
	var b strings.Builder
	b.WriteString("You are a research assistant. Synthesize the following search results and knowledge graph facts into a comprehensive research summary.\n\n")
	fmt.Fprintf(&b, "Research Query: %s\n\n", query)

	b.WriteString("## Search Results\n")
	for i, r := range results {
		fmt.Fprintf(&b, "%d. **%s** (%s)\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet)
	}

	if len(facts.Facts) > 0 {
		b.WriteString("## Knowledge Graph Facts\n")
		for _, f := range facts.Facts {
			fmt.Fprintf(&b, "- %s: %s\n", f.Name, f.FactText)
		}
		b.WriteString("\n")
	}

	b.WriteString("Provide a well-structured markdown summary with sections, key findings, and connections between topics. Include code examples where relevant.")
	return b.String()
}
