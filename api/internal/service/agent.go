package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/urmzd/zoro/api/internal/model"
)

const maxAgentIterations = 10

const systemPrompt = `You are Zoro, an AI research assistant with access to tools. Your purpose is to help users understand topics by searching the web, querying a knowledge graph, and storing important findings.

When answering:
- Use web_search to find current information
- Use search_knowledge to check what's already known
- Use store_knowledge to persist important findings for future reference
- Synthesize information from multiple sources
- Be concise and well-structured in your responses
- Use markdown formatting`

type Agent struct {
	ollama   *OllamaClient
	tools    *ToolRegistry
	registry *ModelRegistry

	mu          sync.RWMutex
	sessions    map[string]*model.ChatSession
	subscribers map[string][]chan model.SSEEvent
}

func NewAgent(ollama *OllamaClient, tools *ToolRegistry, registry *ModelRegistry) *Agent {
	return &Agent{
		ollama:      ollama,
		tools:       tools,
		registry:    registry,
		sessions:    make(map[string]*model.ChatSession),
		subscribers: make(map[string][]chan model.SSEEvent),
	}
}

func (a *Agent) CreateSession() *model.ChatSession {
	session := &model.ChatSession{
		ID:        uuid.New().String(),
		Messages:  []model.ChatMessage{},
		CreatedAt: time.Now(),
	}
	a.mu.Lock()
	a.sessions[session.ID] = session
	a.mu.Unlock()
	return session
}

func (a *Agent) GetSession(id string) (*model.ChatSession, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	s, ok := a.sessions[id]
	return s, ok
}

func (a *Agent) Subscribe(id string) (<-chan model.SSEEvent, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.sessions[id]; !ok {
		return nil, false
	}
	ch := make(chan model.SSEEvent, 128)
	a.subscribers[id] = append(a.subscribers[id], ch)
	return ch, true
}

func (a *Agent) emit(sessionID string, evt model.SSEEvent) {
	a.mu.RLock()
	subs := a.subscribers[sessionID]
	a.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- evt:
		default:
		}
	}
}

func (a *Agent) closeSubscribers(sessionID string) {
	a.mu.Lock()
	subs := a.subscribers[sessionID]
	delete(a.subscribers, sessionID)
	a.mu.Unlock()
	for _, ch := range subs {
		close(ch)
	}
}

func (a *Agent) SendMessage(ctx context.Context, sessionID string, content string) {
	defer a.closeSubscribers(sessionID)

	// Append user message
	userMsg := model.ChatMessage{Role: "user", Content: content}
	a.mu.Lock()
	session, ok := a.sessions[sessionID]
	if ok {
		session.Messages = append(session.Messages, userMsg)
	}
	a.mu.Unlock()

	if !ok {
		a.emit(sessionID, model.SSEEvent{Type: model.EventError, Data: map[string]string{"message": "session not found"}})
		return
	}

	// Clone tool registry and bind store_knowledge to this session
	tools := a.tools.Clone()
	tools.SetStoreKnowledge(sessionID)

	a.runLoop(ctx, sessionID, tools)
}

func (a *Agent) runLoop(ctx context.Context, sessionID string, tools *ToolRegistry) {
	for iteration := 0; iteration < maxAgentIterations; iteration++ {
		// Build Ollama messages from session history
		ollamaMessages := a.buildOllamaMessages(sessionID)

		chunkCh, err := a.ollama.ChatStream(ctx, ollamaMessages, tools.Definitions())
		if err != nil {
			log.Printf("agent ollama error: %v", err)
			a.emit(sessionID, model.SSEEvent{Type: model.EventError, Data: map[string]string{"message": err.Error()}})
			return
		}

		// Accumulate the full assistant response from chunks
		var contentBuf string
		var toolCalls []model.OllamaToolCall

		for chunk := range chunkCh {
			if chunk.Message.Content != "" {
				contentBuf += chunk.Message.Content
				a.emit(sessionID, model.SSEEvent{
					Type: model.EventTextDelta,
					Data: map[string]string{"content": chunk.Message.Content},
				})
			}
			if len(chunk.Message.ToolCalls) > 0 {
				toolCalls = append(toolCalls, chunk.Message.ToolCalls...)
			}
		}

		// Append assistant message to session
		assistantMsg := model.ChatMessage{
			Role:    "assistant",
			Content: contentBuf,
		}

		if len(toolCalls) > 0 {
			for _, tc := range toolCalls {
				argsJSON, _ := json.Marshal(tc.Function.Arguments)
				assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, model.ToolCall{
					ID:        uuid.New().String(),
					Name:      tc.Function.Name,
					Arguments: string(argsJSON),
				})
			}
		}

		a.mu.Lock()
		a.sessions[sessionID].Messages = append(a.sessions[sessionID].Messages, assistantMsg)
		a.mu.Unlock()

		// If no tool calls, we're done
		if len(toolCalls) == 0 {
			break
		}

		// Execute tool calls and append results
		for i, tc := range toolCalls {
			argsJSON, _ := json.Marshal(tc.Function.Arguments)
			callID := assistantMsg.ToolCalls[i].ID

			a.emit(sessionID, model.SSEEvent{
				Type: model.EventToolCallStart,
				Data: map[string]string{
					"id":        callID,
					"name":      tc.Function.Name,
					"arguments": string(argsJSON),
				},
			})

			result, execErr := tools.ExecuteMap(ctx, tc.Function.Name, tc.Function.Arguments)
			if execErr != nil {
				result = fmt.Sprintf("Error: %v", execErr)
			}

			// Update the tool call result in session
			a.mu.Lock()
			msgs := a.sessions[sessionID].Messages
			lastMsg := &msgs[len(msgs)-1]
			if i < len(lastMsg.ToolCalls) {
				lastMsg.ToolCalls[i].Result = result
			}
			a.mu.Unlock()

			a.emit(sessionID, model.SSEEvent{
				Type: model.EventToolCallResult,
				Data: map[string]string{
					"id":     callID,
					"name":   tc.Function.Name,
					"result": result,
				},
			})

			// Append tool result message to session for next iteration
			toolMsg := model.ChatMessage{
				Role:    "tool",
				Content: result,
			}
			a.mu.Lock()
			a.sessions[sessionID].Messages = append(a.sessions[sessionID].Messages, toolMsg)
			a.mu.Unlock()
		}

	}

	a.emit(sessionID, model.SSEEvent{Type: model.EventDone, Data: nil})
}

// ClassifyIntent uses the LLM to determine user intent from a query.
// Returns "chat" for research/conversation or "knowledge_search" for lookups.
func (a *Agent) ClassifyIntent(ctx context.Context, query string) (string, error) {
	prompt := `Classify the user's intent. Reply with ONLY one word: "chat" or "knowledge_search".

Use "knowledge_search" if the user wants to look up, recall, or find something already stored in the knowledge graph.
Use "chat" if the user wants to research something new, have a conversation, or ask a general question.

User query: ` + query

	resp, err := a.ollama.GenerateWithModel(ctx, prompt, a.registry.Model(TierFast))
	if err != nil {
		return "chat", err
	}

	cleaned := strings.TrimSpace(strings.ToLower(resp))
	if strings.Contains(cleaned, "knowledge_search") {
		return "knowledge_search", nil
	}
	return "chat", nil
}

func (a *Agent) buildOllamaMessages(sessionID string) []model.OllamaChatMessage {
	a.mu.RLock()
	session := a.sessions[sessionID]
	msgs := make([]model.ChatMessage, len(session.Messages))
	copy(msgs, session.Messages)
	a.mu.RUnlock()

	var ollamaMsgs []model.OllamaChatMessage

	// System message first
	ollamaMsgs = append(ollamaMsgs, model.OllamaChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, m := range msgs {
		om := model.OllamaChatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		// Re-attach tool calls to assistant messages for Ollama context
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				var args map[string]any
				json.Unmarshal([]byte(tc.Arguments), &args)
				om.ToolCalls = append(om.ToolCalls, model.OllamaToolCall{
					Function: model.OllamaToolCallFunction{
						Name:      tc.Name,
						Arguments: args,
					},
				})
			}
		}
		ollamaMsgs = append(ollamaMsgs, om)
	}

	return ollamaMsgs
}
