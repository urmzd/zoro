package model

type EpisodeRequest struct {
	Name    string `json:"name"`
	Body    string `json:"episode_body"`
	Source  string `json:"source_description"`
	GroupID string `json:"group_id"`
}

type EpisodeResponse struct {
	UUID      string     `json:"uuid"`
	Name      string     `json:"name"`
	Entities  []Entity   `json:"entity_nodes,omitempty"`
	Relations []Relation `json:"episodic_edges,omitempty"`
}

type SearchFactsResponse struct {
	Facts []Fact `json:"facts"`
}

type Fact struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Fact       string `json:"fact"`
	SourceNode Entity `json:"source_node"`
	TargetNode Entity `json:"target_node"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Summary string `json:"summary,omitempty"`
}

type GraphEdge struct {
	ID     string  `json:"id"`
	Source string  `json:"source"`
	Target string  `json:"target"`
	Type   string  `json:"type"`
	Fact   string  `json:"fact,omitempty"`
	Weight float64 `json:"weight"`
}

type NodeDetail struct {
	Node      GraphNode   `json:"node"`
	Neighbors []GraphNode `json:"neighbors"`
	Edges     []GraphEdge `json:"edges"`
}
