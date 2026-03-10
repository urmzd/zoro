package agent

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	agentsdk "github.com/urmzd/agent-sdk"
	"github.com/urmzd/agent-sdk/ollama"
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
- Use markdown formatting`

type Agent struct {
	sdkAgent    *agentsdk.Agent
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
	toolReg := agentsdk.NewToolRegistry(webSearch, searchKG, storeKG)
	sdkAgent := agentsdk.NewAgent(agentsdk.AgentConfig{
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

func (a *Agent) CreateSession() (*models.ChatSession, error) {
	id, err := a.events.CreateSession()
	if err != nil {
		return nil, err
	}
	return &models.ChatSession{
		ID:        id,
		Messages:  []models.ChatMessage{},
		CreatedAt: time.Now(),
	}, nil
}

func (a *Agent) GetSession(id string) (*models.ChatSession, error) {
	return a.events.GetSession(id)
}

func (a *Agent) ListSessions() ([]models.ChatSessionSummary, error) {
	return a.events.ListSessions()
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
	if err := a.events.AppendEvent(sessionID, userEvent); err != nil {
		a.emit(sessionID, models.SSEEvent{
			Type: models.EventError,
			Data: map[string]string{"message": err.Error()},
		})
		a.closeSubscribers(sessionID)
		return
	}

	// Build SDK messages from session history
	session, err := a.events.GetSession(sessionID)
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
	searchKGScoped := a.searchKG.WithGroupID(sessionID)
	storeKGScoped := a.storeKG.WithGroupID(sessionID)
	toolReg := agentsdk.NewToolRegistry(a.webSearch, searchKGScoped, storeKGScoped)

	sessionAgent := agentsdk.NewAgent(agentsdk.AgentConfig{
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

	for delta := range stream.Deltas() {
		switch v := delta.(type) {
		case agentsdk.TextContentDelta:
			contentBuilder.WriteString(v.Content)
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventTextDelta,
				Data: map[string]string{"content": v.Content},
			})
		case agentsdk.ToolCallStartDelta:
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventToolCallStart,
				Data: map[string]string{"id": v.ID, "name": v.Name},
			})
		case agentsdk.ToolCallEndDelta:
			result, _ := v.Arguments["result"].(string)
			id, _ := v.Arguments["id"].(string)
			name, _ := v.Arguments["name"].(string)
			a.emit(sessionID, models.SSEEvent{
				Type: models.EventToolCallResult,
				Data: map[string]string{"id": id, "name": name, "result": result},
			})
			argsJSON, _ := json.Marshal(v.Arguments)
			eventToolCalls = append(eventToolCalls, models.ToolCall{
				ID:   id,
				Name: name,
				Arguments: string(argsJSON),
			})
		case agentsdk.DoneDelta:
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
		a.events.AppendEvent(sessionID, assistantEvent)
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

func buildSDKMessages(msgs []models.ChatMessage) []agentsdk.Message {
	sdkMsgs := make([]agentsdk.Message, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "user":
			sdkMsgs = append(sdkMsgs, agentsdk.NewUserMessage(m.Content))
		case "assistant":
			content := make([]agentsdk.AssistantContent, 0)
			if m.Content != "" {
				content = append(content, agentsdk.TextContent{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				var args map[string]any
				json.Unmarshal([]byte(tc.Arguments), &args)
				content = append(content, agentsdk.ToolUseContent{
					ID:        tc.ID,
					Name:      tc.Name,
					Arguments: args,
				})
			}
			sdkMsgs = append(sdkMsgs, agentsdk.AssistantMessage{Content: content})
		case "tool":
			sdkMsgs = append(sdkMsgs, agentsdk.NewToolResultMessage("", m.Content))
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
	json.Unmarshal([]byte(raw[start:end+1]), &result)
	return result
}
