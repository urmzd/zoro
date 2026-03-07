package model

import "time"

// Chat session and message types

type ChatSession struct {
	ID        string        `json:"id"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
}

type ChatMessage struct {
	Role      string     `json:"role"` // "user", "assistant", "tool"
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Result    string `json:"result,omitempty"`
}

type ChatSessionSummary struct {
	ID           string    `json:"id"`
	Preview      string    `json:"preview"`
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// Ollama /api/chat wire format types

type OllamaChatMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
}

type OllamaToolCall struct {
	Function OllamaToolCallFunction `json:"function"`
}

type OllamaToolCallFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type OllamaTool struct {
	Type     string             `json:"type"`
	Function OllamaToolFunction `json:"function"`
}

type OllamaToolFunction struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Parameters  OllamaToolFunctionParams   `json:"parameters"`
}

type OllamaToolFunctionParams struct {
	Type       string                        `json:"type"`
	Required   []string                      `json:"required"`
	Properties map[string]OllamaToolProperty `json:"properties"`
}

type OllamaToolProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type OllamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []OllamaChatMessage `json:"messages"`
	Tools    []OllamaTool        `json:"tools,omitempty"`
	Stream   bool                `json:"stream"`
}

type OllamaChatChunk struct {
	Message OllamaChatMessage `json:"message"`
	Done    bool              `json:"done"`
}
