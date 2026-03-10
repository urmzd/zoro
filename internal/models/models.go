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

// ── SSE event types ─────────────────────────────────────────────────

const (
	EventSearchStarted    = "search_started"
	EventSearchResults    = "search_results"
	EventEpisodeIngested  = "episode_ingested"
	EventEntityDiscovered = "entity_discovered"
	EventRelationFound    = "relation_found"
	EventPriorKnowledge   = "prior_knowledge"
	EventGraphReady       = "graph_ready"
	EventSummaryToken     = "summary_token"
	EventResearchComplete = "research_complete"
	EventError            = "error"
	EventTextDelta        = "text_delta"
	EventToolCallStart    = "tool_call_start"
	EventToolCallResult   = "tool_call_result"
	EventDone             = "done"
)

type SSEEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// ── Research types ──────────────────────────────────────────────────

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type ResearchSession struct {
	ID        string          `json:"id"`
	Query     string          `json:"query"`
	Status    string          `json:"status"`
	Results   []SearchResult  `json:"results"`
	Timeline  []TimelineEvent `json:"timeline"`
	Summary   string          `json:"summary"`
	CreatedAt time.Time       `json:"created_at"`
}

type TimelineEvent struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// ── Intent / Autocomplete response types ────────────────────────────

type IntentResponse struct {
	Action string `json:"action"`
	Query  string `json:"query"`
}

type AutocompleteResponse struct {
	Suggestions []string `json:"suggestions"`
}
