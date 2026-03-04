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
)

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data" swaggertype:"object"`
}
