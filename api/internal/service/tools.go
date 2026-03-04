package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urmzd/zoro/api/internal/model"
)

// ToolFunc executes a tool and returns a concise text result.
type ToolFunc func(ctx context.Context, args map[string]any) (string, error)

// ToolRegistry holds available tools and their definitions.
type ToolRegistry struct {
	searcher  *Searcher
	knowledge KnowledgeStore

	// Per-session state for store_knowledge
	storeGroupID string

	funcs map[string]ToolFunc
	defs  []model.OllamaTool
}

func NewToolRegistry(searcher *Searcher, knowledge KnowledgeStore) *ToolRegistry {
	tr := &ToolRegistry{
		searcher:  searcher,
		knowledge: knowledge,
		funcs:     make(map[string]ToolFunc),
	}
	tr.register()
	return tr
}

// Clone returns a copy that can be bound to a specific session without shared mutable state.
func (tr *ToolRegistry) Clone() *ToolRegistry {
	return &ToolRegistry{
		searcher:  tr.searcher,
		knowledge: tr.knowledge,
		funcs:     tr.funcs,
		defs:      tr.defs,
	}
}

// SetStoreKnowledge binds store_knowledge to a specific groupID (chat session).
func (tr *ToolRegistry) SetStoreKnowledge(groupID string) {
	tr.storeGroupID = groupID
}

func (tr *ToolRegistry) Definitions() []model.OllamaTool {
	return tr.defs
}

func (tr *ToolRegistry) Execute(ctx context.Context, name string, argsJSON string) (string, error) {
	fn, ok := tr.funcs[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	var args map[string]any
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("parse tool args: %w", err)
		}
	}
	if args == nil {
		args = make(map[string]any)
	}

	return fn(ctx, args)
}

func (tr *ToolRegistry) ExecuteMap(ctx context.Context, name string, args map[string]any) (string, error) {
	fn, ok := tr.funcs[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	if args == nil {
		args = make(map[string]any)
	}
	return fn(ctx, args)
}

func (tr *ToolRegistry) register() {
	// web_search
	tr.funcs["web_search"] = tr.webSearch
	tr.defs = append(tr.defs, model.OllamaTool{
		Type: "function",
		Function: model.OllamaToolFunction{
			Name:        "web_search",
			Description: "Search the web for current information on a topic. Returns up to 5 results with titles, URLs, and snippets.",
			Parameters: model.OllamaToolFunctionParams{
				Type:     "object",
				Required: []string{"query"},
				Properties: map[string]model.OllamaToolProperty{
					"query": {Type: "string", Description: "The search query"},
				},
			},
		},
	})

	// search_knowledge
	tr.funcs["search_knowledge"] = tr.searchKnowledge
	tr.defs = append(tr.defs, model.OllamaTool{
		Type: "function",
		Function: model.OllamaToolFunction{
			Name:        "search_knowledge",
			Description: "Search the knowledge graph for previously stored facts and entities. Returns up to 10 relevant facts.",
			Parameters: model.OllamaToolFunctionParams{
				Type:     "object",
				Required: []string{"query"},
				Properties: map[string]model.OllamaToolProperty{
					"query": {Type: "string", Description: "The search query for knowledge retrieval"},
				},
			},
		},
	})

	// store_knowledge
	tr.funcs["store_knowledge"] = tr.storeKnowledge
	tr.defs = append(tr.defs, model.OllamaTool{
		Type: "function",
		Function: model.OllamaToolFunction{
			Name:        "store_knowledge",
			Description: "Store information into the knowledge graph by extracting entities and relationships from text. Use this to persist important findings.",
			Parameters: model.OllamaToolFunctionParams{
				Type:     "object",
				Required: []string{"text", "source"},
				Properties: map[string]model.OllamaToolProperty{
					"text":   {Type: "string", Description: "The text content to extract knowledge from"},
					"source": {Type: "string", Description: "Description of the source of this information"},
				},
			},
		},
	})
}

func (tr *ToolRegistry) webSearch(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("web_search: query is required")
	}

	results, err := tr.searcher.Search(ctx, query)
	if err != nil {
		return "", fmt.Errorf("web_search: %w", err)
	}

	// Cap at 5 results with truncated snippets
	var b strings.Builder
	limit := 5
	if len(results) < limit {
		limit = len(results)
	}
	for i := 0; i < limit; i++ {
		r := results[i]
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

func (tr *ToolRegistry) searchKnowledge(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("search_knowledge: query is required")
	}

	resp, err := tr.knowledge.SearchFacts(ctx, query, "")
	if err != nil {
		return "", fmt.Errorf("search_knowledge: %w", err)
	}

	// Cap at 10 facts
	var b strings.Builder
	limit := 10
	if len(resp.Facts) < limit {
		limit = len(resp.Facts)
	}
	for i := 0; i < limit; i++ {
		f := resp.Facts[i]
		fmt.Fprintf(&b, "- %s -> %s: %s\n", f.SourceNode.Name, f.TargetNode.Name, f.Fact)
	}

	if b.Len() == 0 {
		return "No relevant knowledge found.", nil
	}
	return b.String(), nil
}

func (tr *ToolRegistry) storeKnowledge(ctx context.Context, args map[string]any) (string, error) {
	text, _ := args["text"].(string)
	source, _ := args["source"].(string)
	if text == "" {
		return "", fmt.Errorf("store_knowledge: text is required")
	}
	if source == "" {
		source = "chat"
	}

	groupID := tr.storeGroupID

	resp, err := tr.knowledge.AddEpisode(ctx, model.EpisodeRequest{
		Name:    source,
		Body:    text,
		Source:  source,
		GroupID: groupID,
	})
	if err != nil {
		return "", fmt.Errorf("store_knowledge: %w", err)
	}

	return fmt.Sprintf("Stored %d entities and %d relations.", len(resp.Entities), len(resp.Relations)), nil
}
