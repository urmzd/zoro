package tools

import (
	"context"
	"fmt"
	"strings"

	agentsdk "github.com/urmzd/agent-sdk"
	"github.com/urmzd/zoro/internal/searcher"
)

// WebSearchTool implements agentsdk.Tool for web searching.
type WebSearchTool struct {
	searcher *searcher.Searcher
}

func NewWebSearchTool(s *searcher.Searcher) *WebSearchTool {
	return &WebSearchTool{searcher: s}
}

func (t *WebSearchTool) Definition() agentsdk.ToolDef {
	return agentsdk.ToolDef{
		Name:        "web_search",
		Description: "Search the web for current information on a topic. Returns up to 5 results with titles, URLs, and snippets.",
		Parameters: agentsdk.ParameterSchema{
			Type:     "object",
			Required: []string{"query"},
			Properties: map[string]agentsdk.PropertyDef{
				"query": {Type: "string", Description: "The search query"},
			},
		},
	}
}

func (t *WebSearchTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("web_search: query is required")
	}

	results, err := t.searcher.Search(ctx, query)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	limit := len(results)
	if limit > 5 {
		limit = 5
	}
	for i, r := range results[:limit] {
		snippet := r.Snippet
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		fmt.Fprintf(&b, "%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, snippet)
	}

	if b.Len() == 0 {
		return "No results found.", nil
	}
	return b.String(), nil
}
