package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/urmzd/zoro/api/internal/model"
)

type Orchestrator struct {
	knowledge KnowledgeStore
	ollama    *OllamaClient
	searcher  *Searcher

	mu          sync.RWMutex
	sessions    map[string]*model.ResearchSession
	subscribers map[string][]chan model.SSEEvent
}

func NewOrchestrator(knowledge KnowledgeStore, ollama *OllamaClient, searcher *Searcher) *Orchestrator {
	return &Orchestrator{
		knowledge:   knowledge,
		ollama:      ollama,
		searcher:    searcher,
		sessions:    make(map[string]*model.ResearchSession),
		subscribers: make(map[string][]chan model.SSEEvent),
	}
}

func (o *Orchestrator) CreateSession(query string) *model.ResearchSession {
	session := &model.ResearchSession{
		ID:        uuid.New().String(),
		Query:     query,
		Status:    "created",
		Results:   []model.SearchResult{},
		Entities:  []model.Entity{},
		Relations: []model.Relation{},
		Timeline:  []model.TimelineEvent{},
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	o.sessions[session.ID] = session
	o.mu.Unlock()

	return session
}

func (o *Orchestrator) GetSession(id string) (*model.ResearchSession, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	s, ok := o.sessions[id]
	return s, ok
}

func (o *Orchestrator) Subscribe(id string) (<-chan model.SSEEvent, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, ok := o.sessions[id]; !ok {
		return nil, false
	}
	ch := make(chan model.SSEEvent, 128)
	o.subscribers[id] = append(o.subscribers[id], ch)
	return ch, true
}

func (o *Orchestrator) emit(sessionID string, evt model.SSEEvent) {
	o.mu.Lock()
	session := o.sessions[sessionID]
	if session != nil {
		session.Timeline = append(session.Timeline, model.TimelineEvent{
			Type:      evt.Type,
			Message:   fmt.Sprintf("%v", evt.Data),
			Timestamp: time.Now(),
		})
	}
	subs := o.subscribers[sessionID]
	o.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- evt:
		default:
		}
	}
}

func (o *Orchestrator) closeSubscribers(sessionID string) {
	o.mu.Lock()
	subs := o.subscribers[sessionID]
	delete(o.subscribers, sessionID)
	o.mu.Unlock()
	for _, ch := range subs {
		close(ch)
	}
}

func (o *Orchestrator) setStatus(sessionID, status string) {
	o.mu.Lock()
	if s, ok := o.sessions[sessionID]; ok {
		s.Status = status
	}
	o.mu.Unlock()
}

func (o *Orchestrator) Run(ctx context.Context, sessionID string) {
	defer o.closeSubscribers(sessionID)

	o.setStatus(sessionID, "running")

	session, ok := o.GetSession(sessionID)
	if !ok {
		return
	}
	query := session.Query

	// Step 1: Query knowledge store for prior knowledge
	o.emit(sessionID, model.SSEEvent{Type: model.EventPriorKnowledge, Data: map[string]string{"message": "Searching prior knowledge..."}})
	priorFacts, err := o.knowledge.SearchFacts(ctx, query, "")
	if err != nil {
		log.Printf("prior knowledge search error (non-fatal): %v", err)
	} else if len(priorFacts.Facts) > 0 {
		o.emit(sessionID, model.SSEEvent{Type: model.EventPriorKnowledge, Data: priorFacts.Facts})
	}

	// Step 2: Web search
	o.emit(sessionID, model.SSEEvent{Type: model.EventSearchStarted, Data: map[string]string{"query": query}})
	results, err := o.searcher.Search(ctx, query)
	if err != nil {
		o.emit(sessionID, model.SSEEvent{Type: model.EventError, Data: map[string]string{"error": err.Error()}})
		o.setStatus(sessionID, "error")
		return
	}

	o.mu.Lock()
	o.sessions[sessionID].Results = results
	o.mu.Unlock()

	o.emit(sessionID, model.SSEEvent{Type: model.EventSearchResults, Data: results})

	// Step 3: Ingest each result into knowledge store as an episode
	for i, result := range results {
		episodeBody := fmt.Sprintf("Title: %s\nURL: %s\nSnippet: %s", result.Title, result.URL, result.Snippet)
		episodeReq := model.EpisodeRequest{
			Name:    fmt.Sprintf("%s - Result %d", query, i+1),
			Body:    episodeBody,
			Source:  result.URL,
			GroupID: sessionID,
		}

		resp, err := o.knowledge.AddEpisode(ctx, episodeReq)
		if err != nil {
			log.Printf("episode ingestion error (result %d): %v", i+1, err)
			continue
		}

		o.emit(sessionID, model.SSEEvent{Type: model.EventEpisodeIngested, Data: map[string]interface{}{
			"result_index": i,
			"episode_uuid": resp.UUID,
		}})

		// Emit discovered entities
		if resp.Entities != nil {
			for _, entity := range resp.Entities {
				o.mu.Lock()
				o.sessions[sessionID].Entities = append(o.sessions[sessionID].Entities, entity)
				o.mu.Unlock()
				o.emit(sessionID, model.SSEEvent{Type: model.EventEntityDiscovered, Data: entity})
			}
		}

		// Emit discovered relations
		if resp.Relations != nil {
			for _, relation := range resp.Relations {
				o.mu.Lock()
				o.sessions[sessionID].Relations = append(o.sessions[sessionID].Relations, relation)
				o.mu.Unlock()
				o.emit(sessionID, model.SSEEvent{Type: model.EventRelationFound, Data: relation})
			}
		}
	}

	// Step 4: Get session subgraph from knowledge store
	subgraph, err := o.knowledge.SearchFacts(ctx, query, sessionID)
	if err != nil {
		log.Printf("subgraph search error: %v", err)
	} else {
		o.emit(sessionID, model.SSEEvent{Type: model.EventGraphReady, Data: subgraph})
	}

	// Step 5: Stream LLM synthesis via Ollama
	summaryParts := buildSummaryPrompt(query, results, subgraph)
	tokenCh, err := o.ollama.GenerateStream(ctx, summaryParts)
	if err != nil {
		log.Printf("ollama stream error: %v", err)
	} else {
		var summaryBuilder strings.Builder
		for token := range tokenCh {
			summaryBuilder.WriteString(token)
			o.emit(sessionID, model.SSEEvent{Type: model.EventSummaryToken, Data: map[string]string{"token": token}})
		}
		o.mu.Lock()
		o.sessions[sessionID].Summary = summaryBuilder.String()
		o.mu.Unlock()
	}

	// Step 6: Complete
	o.setStatus(sessionID, "complete")
	o.emit(sessionID, model.SSEEvent{Type: model.EventResearchComplete, Data: map[string]string{"session_id": sessionID}})
}

func buildSummaryPrompt(query string, results []model.SearchResult, facts *model.SearchFactsResponse) string {
	var b strings.Builder
	b.WriteString("You are a research assistant. Synthesize the following search results and knowledge graph facts into a comprehensive research summary.\n\n")
	b.WriteString(fmt.Sprintf("Research Query: %s\n\n", query))

	b.WriteString("## Search Results\n")
	for i, r := range results {
		b.WriteString(fmt.Sprintf("%d. **%s** (%s)\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet))
	}

	if facts != nil && len(facts.Facts) > 0 {
		b.WriteString("## Knowledge Graph Facts\n")
		for _, f := range facts.Facts {
			b.WriteString(fmt.Sprintf("- %s: %s\n", f.Name, f.Fact))
		}
		b.WriteString("\n")
	}

	b.WriteString("Provide a well-structured markdown summary with sections, key findings, and connections between topics. Include code examples where relevant.")
	return b.String()
}
