package model

import "time"

type ResearchSession struct {
	ID        string          `json:"id"`
	Query     string          `json:"query"`
	Status    string          `json:"status"`
	Results   []SearchResult  `json:"results"`
	Entities  []Entity        `json:"entities"`
	Relations []Relation      `json:"relations"`
	Timeline  []TimelineEvent `json:"timeline"`
	Summary   string          `json:"summary"`
	CreatedAt time.Time       `json:"created_at"`
}

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type Entity struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

type Relation struct {
	UUID       string `json:"uuid"`
	SourceUUID string `json:"source_uuid"`
	TargetUUID string `json:"target_uuid"`
	Type       string `json:"type"`
	Fact       string `json:"fact"`
}

type TimelineEvent struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty" swaggertype:"object"`
}
