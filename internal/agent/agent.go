package agent

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/urmzd/adk"
	"github.com/urmzd/adk/core"
	"github.com/urmzd/adk/provider/ollama"
	"github.com/urmzd/zoro/internal/events"
	"github.com/urmzd/zoro/internal/models"
	"github.com/urmzd/zoro/internal/tools"
)

const systemPrompt = `You are Zoro, an AI research assistant with access to tools. Your purpose is to help users understand topics by searching the web, querying a knowledge graph, and storing important findings.

When answering:
- Use web_search to find current information
- Use search_knowledge to check what's already known
- Use store_knowledge to persist important findings for future reference
- Synthesize information from multiple sources
- Be concise and well-structured in your responses
- Use markdown formatting
- IMPORTANT: When citing information from web_search results, always include inline citations using the result index numbers like [1], [2], [3]. For example: "React 19 introduced server components [1] and improved hydration [3]." Every factual claim from search results must have a citation.`

type Agent struct {
	sdkAgent    *adk.Agent
	adapter     *ollama.Adapter
	fastModel   string
	events      *events.Store
	webSearch   *tools.WebSearchTool
	searchKG    *tools.SearchKnowledgeTool
	storeKG     *tools.StoreKnowledgeTool
	subscribers map[string][]chan models.SSEEvent
	mu          sync.RWMutex
}

func New(adapter *ollama.Adapter, webSearch *tools.WebSearchTool, searchKG *tools.SearchKnowledgeTool, storeKG *tools.StoreKnowledgeTool, fastModel string, e *events.Store) *Agent {
	toolReg := core.NewToolRegistry(webSearch, searchKG, storeKG)
	sdkAgent := adk.NewAgent(adk.AgentConfig{
		Name:         "zoro",
		SystemPrompt: systemPrompt,
		Provider:     adapter,
		Tools:        toolReg,
		MaxIter:      10,
	})

	return &Agent{
		sdkAgent:    sdkAgent,
		adapter:     adapter,
		fastModel:   fastModel,
		events:      e,
		webSearch:   webSearch,
		searchKG:    searchKG,
		storeKG:     storeKG,
		subscribers: make(map[string][]chan models.SSEEvent),
	}
}

func (a *Agent) CreateSession(ctx context.Context) (*models.ChatSession, error) {
	id, err := a.events.CreateSession(ctx)
	if err != nil {
		return nil, err
	}
	return &models.ChatSession{
		ID:        id,
		Messages:  []models.ChatMessage{},
		CreatedAt: time.Now(),
	}, nil
}

func (a *Agent) GetSession(ctx context.Context, id string) (*models.ChatSession, error) {
	return a.events.GetSession(ctx, id)
}

func (a *Agent) ListSessions(ctx context.Context) ([]models.ChatSessionSummary, error) {
	return a.events.ListSessions(ctx)
}

func (a *Agent) Subscribe(id string) <-chan models.SSEEvent {
	ch := make(chan models.SSEEvent, 128)
	a.mu.Lock()
	a.subscribers[id] = append(a.subscribers[id], ch)
	a.mu.Unlock()
	return ch
}

func (a *Agent) emit(sessionID string, evt models.SSEEvent) {
	a.mu.RLock()
	channels := a.subscribers[sessionID]
	a.mu.RUnlock()
	for _, ch := range channels {
		select {
		case ch <- evt:
		default:
			log.Printf("[agent] warning: dropped event %s for session %s (slow subscriber)", evt.Type, sessionID)
		}
	}
}

func (a *Agent) closeSubscribers(sessionID string) {
	a.mu.Lock()
	channels := a.subscribers[sessionID]
	delete(a.subscribers, sessionID)
	a.mu.Unlock()
	for _, ch := range channels {
		close(ch)
	}
}

func (a *Agent) SendMessage(ctx context.Context, sessionID, content string) {
	userEvent := models.ChatEvent{
		SessionID: sessionID,
		Type:      "user_message",
		Role:      "user",
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := a.events.AppendEvent(ctx, sessionID, userEvent); err != nil {
		a.emit(sessionID, models.SSEEvent{
			Type: models.EventError,
			Data: map[string]string{"message": err.Error()},
		})
		a.closeSubscribers(sessionID)
		return
	}

	// Build SDK messages from session history
	session, err := a.events.GetSession(ctx, sessionID)
	if err != nil {
		a.emit(sessionID, models.SSEEvent{
			Type: models.EventError,
			Data: map[string]string{"message": err.Error()},
		})
		a.closeSubscribers(sessionID)
		return
	}

	sdkMessages := buildSDKMessages(session.Messages)

	// Create per-session tools with group ID
	webSearchScoped := a.webSearch.WithGroupID(sessionID)
	searchKGScoped := a.searchKG.WithGroupID(sessionID)
	storeKGScoped := a.storeKG.WithGroupID(sessionID)
	toolReg := core.NewToolRegistry(webSearchScoped, searchKGScoped, storeKGScoped)

	sessionAgent := adk.NewAgent(adk.AgentConfig{
		Name:         "zoro",
		SystemPrompt: systemPrompt,
		Provider:     a.adapter,
		Tools:        toolReg,
		MaxIter:      10,
	})

	stream := sessionAgent.Invoke(ctx, sdkMessages)

	// Forward deltas to SSE subscribers and persist
	var contentBuilder strings.Builder
	var eventToolCalls []models.ToolCall

	// Track active tool calls by ID for correlating start/end deltas.
	// lastStartedID tracks sequential ordering: the SDK emits
	// ToolCallStartDelta then ToolCallEndDelta for each call in order.
	type activeCall struct {
		ID   string
		Name string
		Args map[string]any
	}
	activeCalls := make(map[string]*activeCall)
	var lastStartedID string

	for delta := range stream.Deltas() {
		switch v := delta.(type) {
		case core.TextContentDelta:
			contentBuilder.WriteString(v.Content)
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventTextDelta,
				Data: map[string]string{"content": v.Content},
			})
		case core.ToolCallStartDelta:
			activeCalls[v.ID] = &activeCall{ID: v.ID, Name: v.Name}
			lastStartedID = v.ID
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventToolCallStart,
				Data: map[string]string{"id": v.ID, "name": v.Name},
			})
		case core.ToolCallEndDelta:
			// ToolCallEndDelta carries the LLM's parsed arguments but has no ID.
			// Match by sequential ordering (same as SDK's DefaultAggregator).
			if lastStartedID != "" {
				if ac, ok := activeCalls[lastStartedID]; ok {
					ac.Args = v.Arguments
				}
				lastStartedID = ""
			}
		case core.ToolExecEndDelta:
			// ToolExecEndDelta carries the actual tool result.
			ac := activeCalls[v.ToolCallID]
			name := ""
			if ac != nil {
				name = ac.Name
			}
			result := v.Result
			if v.Error != "" {
				result = "error: " + v.Error
			}
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventToolCallResult,
				Data: map[string]string{"id": v.ToolCallID, "name": name, "result": result},
			})
			if ac != nil {
				argsJSON, _ := json.Marshal(ac.Args)
				eventToolCalls = append(eventToolCalls, models.ToolCall{
					ID:        ac.ID,
					Name:      ac.Name,
					Arguments: string(argsJSON),
				})
				delete(activeCalls, v.ToolCallID)
			}
		case core.ErrorDelta:
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventError,
				Data: map[string]string{"message": v.Error.Error()},
			})
		case core.DoneDelta:
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventDone,
				Data: nil,
			})
		}
	}

	// Persist assistant response
	if contentBuilder.Len() > 0 || len(eventToolCalls) > 0 {
		assistantEvent := models.ChatEvent{
			SessionID: sessionID,
			Type:      "assistant_message",
			Role:      "assistant",
			Content:   contentBuilder.String(),
			ToolCalls: eventToolCalls,
			CreatedAt: time.Now(),
		}
		if err := a.events.AppendEvent(ctx, sessionID, assistantEvent); err != nil {
			log.Printf("[agent] warning: failed to persist assistant message for session %s: %v", sessionID, err)
		}
	}

	a.closeSubscribers(sessionID)
}

func (a *Agent) ClassifyIntent(ctx context.Context, query string) (string, error) {
	prompt := `Classify the user's intent as "chat" or "knowledge_search".

"knowledge_search": user wants to look up or find something stored in the knowledge graph.
"chat": user wants to research something new, have a conversation, or ask a general question.

User query: ` + query

	format := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
				"enum": []string{"chat", "knowledge_search"},
			},
		},
		"required": []string{"action"},
	}

	resp, err := a.adapter.GenerateWithModel(ctx, prompt, a.fastModel, format, nil)
	if err != nil {
		return "chat", err
	}

	var result struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil || result.Action != "knowledge_search" {
		return "chat", nil
	}
	return "knowledge_search", nil
}

func (a *Agent) Autocomplete(ctx context.Context, query string) []string {
	prompt := `Given the partial query below, suggest 3 to 5 complete search queries the user might intend. Return ONLY a JSON array of strings, no extra text.

Partial query: ` + query

	raw, err := a.adapter.GenerateWithModel(ctx, prompt, a.fastModel, nil, nil)
	if err != nil {
		return nil
	}

	return parseJSONArray(raw)
}

func buildSDKMessages(msgs []models.ChatMessage) []core.Message {
	sdkMsgs := make([]core.Message, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "user":
			sdkMsgs = append(sdkMsgs, core.NewUserMessage(m.Content))
		case "assistant":
			content := make([]core.AssistantContent, 0)
			if m.Content != "" {
				content = append(content, core.TextContent{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
					log.Printf("[agent] warning: failed to parse tool call arguments for %s: %v", tc.Name, err)
				}
				content = append(content, core.ToolUseContent{
					ID:        tc.ID,
					Name:      tc.Name,
					Arguments: args,
				})
			}
			sdkMsgs = append(sdkMsgs, core.AssistantMessage{Content: content})
		case "tool":
			sdkMsgs = append(sdkMsgs, core.NewToolResultMessage(core.ToolResultContent{Text: m.Content}))
		}
	}
	return sdkMsgs
}

func parseJSONArray(raw string) []string {
	start := strings.Index(raw, "[")
	if start < 0 {
		return nil
	}
	end := strings.LastIndex(raw, "]")
	if end <= start {
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		log.Printf("[agent] warning: failed to parse autocomplete JSON: %v", err)
		return nil
	}
	return result
}
