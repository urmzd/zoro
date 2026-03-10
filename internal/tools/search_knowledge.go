package tools

import (
	"context"
	"fmt"
	"strings"

	agentsdk "github.com/urmzd/agent-sdk"
	kg "github.com/urmzd/knowledge-graph-sdk"
)

// SearchKnowledgeTool implements agentsdk.Tool for knowledge graph search.
type SearchKnowledgeTool struct {
	graph   kg.Graph
	groupID string
}

func NewSearchKnowledgeTool(graph kg.Graph) *SearchKnowledgeTool {
	return &SearchKnowledgeTool{graph: graph}
}

func (t *SearchKnowledgeTool) WithGroupID(id string) *SearchKnowledgeTool {
	return &SearchKnowledgeTool{graph: t.graph, groupID: id}
}

func (t *SearchKnowledgeTool) Definition() agentsdk.ToolDef {
	return agentsdk.ToolDef{
		Name:        "search_knowledge",
		Description: "Search the knowledge graph for previously stored facts and entities. Returns up to 10 relevant facts.",
		Parameters: agentsdk.ParameterSchema{
			Type:     "object",
			Required: []string{"query"},
			Properties: map[string]agentsdk.PropertyDef{
				"query": {Type: "string", Description: "The search query for knowledge retrieval"},
			},
		},
	}
}

func (t *SearchKnowledgeTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("search_knowledge: query is required")
	}

	opts := []kg.SearchOption{}
	if t.groupID != "" {
		opts = append(opts, kg.WithGroupID(t.groupID))
	}

	resp, err := t.graph.SearchFacts(ctx, query, opts...)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	limit := len(resp.Facts)
	if limit > 10 {
		limit = 10
	}
	for _, f := range resp.Facts[:limit] {
		fmt.Fprintf(&b, "- %s -> %s: %s\n", f.SourceNode.Name, f.TargetNode.Name, f.FactText)
	}

	if b.Len() == 0 {
		return "No relevant knowledge found.", nil
	}
	return b.String(), nil
}
