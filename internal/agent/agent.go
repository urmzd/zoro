package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	saige "github.com/urmzd/saige/agent"
	"github.com/urmzd/saige/agent/provider/ollama"
	"github.com/urmzd/saige/agent/types"
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
	adapter   *ollama.Adapter
	events    *events.Store
	webSearch *tools.WebSearchTool
	searchKG  *tools.SearchKnowledgeTool
	storeKG   *tools.StoreKnowledgeTool
}

func New(adapter *ollama.Adapter, webSearch *tools.WebSearchTool, searchKG *tools.SearchKnowledgeTool, storeKG *tools.StoreKnowledgeTool, e *events.Store) *Agent {
	return &Agent{
		adapter:   adapter,
		events:    e,
		webSearch: webSearch,
		searchKG:  searchKG,
		storeKG:   storeKG,
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

// InvokeSync sends a message in the given session and returns the full response.
// If sessionID is empty, a new session is created.
func (a *Agent) InvokeSync(ctx context.Context, sessionID, content string) (response string, returnedSessionID string, err error) {
	if sessionID == "" {
		session, err := a.CreateSession(ctx)
		if err != nil {
			return "", "", fmt.Errorf("create session: %w", err)
		}
		sessionID = session.ID
	}

	// Persist user message
	userEvent := models.ChatEvent{
		SessionID: sessionID,
		Type:      "user_message",
		Role:      "user",
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := a.events.AppendEvent(ctx, sessionID, userEvent); err != nil {
		return "", sessionID, fmt.Errorf("persist user message: %w", err)
	}

	// Build SDK messages from session history
	session, err := a.events.GetSession(ctx, sessionID)
	if err != nil {
		return "", sessionID, fmt.Errorf("get session: %w", err)
	}

	sdkMessages := buildSDKMessages(session.Messages)

	// Create per-session tools with group ID
	webSearchScoped := a.webSearch.WithGroupID(sessionID)
	searchKGScoped := a.searchKG.WithGroupID(sessionID)
	storeKGScoped := a.storeKG.WithGroupID(sessionID)
	toolReg := types.NewToolRegistry(webSearchScoped, searchKGScoped, storeKGScoped)

	sessionAgent := saige.NewAgent(saige.AgentConfig{
		Name:         "zoro",
		SystemPrompt: systemPrompt,
		Provider:     a.adapter,
		Tools:        toolReg,
		MaxIter:      10,
	})

	stream := sessionAgent.Invoke(ctx, sdkMessages)

	// Drain deltas and collect response
	var contentBuilder strings.Builder
	var eventToolCalls []models.ToolCall

	type activeCall struct {
		ID   string
		Name string
		Args map[string]any
	}
	activeCalls := make(map[string]*activeCall)
	var lastStartedID string

	for delta := range stream.Deltas() {
		switch v := delta.(type) {
		case types.TextContentDelta:
			contentBuilder.WriteString(v.Content)
		case types.ToolCallStartDelta:
			activeCalls[v.ID] = &activeCall{ID: v.ID, Name: v.Name}
			lastStartedID = v.ID
		case types.ToolCallEndDelta:
			if lastStartedID != "" {
				if ac, ok := activeCalls[lastStartedID]; ok {
					ac.Args = v.Arguments
				}
				lastStartedID = ""
			}
		case types.ToolExecEndDelta:
			ac := activeCalls[v.ToolCallID]
			if ac != nil {
				argsJSON, _ := json.Marshal(ac.Args)
				eventToolCalls = append(eventToolCalls, models.ToolCall{
					ID:        ac.ID,
					Name:      ac.Name,
					Arguments: string(argsJSON),
				})
				delete(activeCalls, v.ToolCallID)
			}
		case types.ErrorDelta:
			log.Printf("[agent] error delta in session %s: %v", sessionID, v.Error)
		case types.DoneDelta:
			// stream complete
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

	return contentBuilder.String(), sessionID, nil
}

func buildSDKMessages(msgs []models.ChatMessage) []types.Message {
	sdkMsgs := make([]types.Message, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "user":
			sdkMsgs = append(sdkMsgs, types.NewUserMessage(m.Content))
		case "assistant":
			content := make([]types.AssistantContent, 0)
			if m.Content != "" {
				content = append(content, types.TextContent{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
					log.Printf("[agent] warning: failed to parse tool call arguments for %s: %v", tc.Name, err)
				}
				content = append(content, types.ToolUseContent{
					ID:        tc.ID,
					Name:      tc.Name,
					Arguments: args,
				})
			}
			sdkMsgs = append(sdkMsgs, types.AssistantMessage{Content: content})
		case "tool":
			sdkMsgs = append(sdkMsgs, types.NewToolResultMessage(types.ToolResultContent{Text: m.Content}))
		}
	}
	return sdkMsgs
}
