package model

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

	// Chat SSE events
	EventTextDelta      = "text_delta"
	EventToolCallStart  = "tool_call_start"
	EventToolCallResult = "tool_call_result"
	EventDone           = "done"
)

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data" swaggertype:"object"`
}
