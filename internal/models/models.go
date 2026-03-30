package models

import "time"

// ── Chat types ──────────────────────────────────────────────────────

type ChatSession struct {
	ID        string        `json:"id"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
}

type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"toolCalls,omitempty"`
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
	MessageCount int64     `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type ChatEvent struct {
	ID        string     `json:"id"`
	SessionID string     `json:"session_id"`
	Type      string     `json:"type"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"toolCalls,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ── Research types ──────────────────────────────────────────────────

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}
