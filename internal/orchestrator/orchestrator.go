package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/urmzd/saige/agent/provider/ollama"
	kgtypes "github.com/urmzd/saige/knowledge/types"
	"github.com/urmzd/zoro/internal/models"
	"github.com/urmzd/zoro/internal/searcher"
)

type Orchestrator struct {
	graph    kgtypes.Graph
	adapter  *ollama.Adapter
	searcher *searcher.Searcher
}

func New(g kgtypes.Graph, a *ollama.Adapter, s *searcher.Searcher) *Orchestrator {
	return &Orchestrator{
		graph:    g,
		adapter:  a,
		searcher: s,
	}
}

// RunSync runs the full research pipeline synchronously and returns the summary.
func (o *Orchestrator) RunSync(ctx context.Context, query string) (string, error) {
	// Step 1: Prior knowledge
	priorFacts, err := o.graph.SearchFacts(ctx, query)
	if err != nil {
		log.Printf("prior knowledge search error (non-fatal): %v", err)
	}

	// Step 2: Web search
	results, err := o.searcher.Search(ctx, query)
	if err != nil {
		return "", fmt.Errorf("web search: %w", err)
	}

	// Step 3: Ingest each result
	groupID := "research-" + query
	for i, result := range results {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		episodeBody := fmt.Sprintf("Title: %s\nURL: %s\nSnippet: %s", result.Title, result.URL, result.Snippet)
		input := &kgtypes.EpisodeInput{
			Name:    fmt.Sprintf("%s - Result %d", query, i+1),
			Body:    episodeBody,
			Source:  result.URL,
			GroupID: groupID,
		}
		if _, err := o.graph.IngestEpisode(ctx, input); err != nil {
			log.Printf("episode ingestion error (result %d): %v", i+1, err)
		}
	}

	// Step 4: Get session subgraph
	subgraph, err := o.graph.SearchFacts(ctx, query, kgtypes.WithGroupID(groupID))
	if err != nil {
		log.Printf("subgraph search error: %v", err)
		subgraph = priorFacts
	}

	if subgraph == nil {
		return "No results found.", nil
	}

	// Step 5: LLM synthesis
	prompt := buildSummaryPrompt(query, results, subgraph)
	rx, err := o.adapter.GenerateStream(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("llm stream: %w", err)
	}

	var summary strings.Builder
	for token := range rx {
		if ctx.Err() != nil {
			return summary.String(), ctx.Err()
		}
		summary.WriteString(token)
	}

	return summary.String(), nil
}

func buildSummaryPrompt(query string, results []models.SearchResult, facts *kgtypes.SearchFactsResult) string {
	var b strings.Builder
	b.WriteString("You are a research assistant. Synthesize the following search results and knowledge graph facts into a comprehensive research summary.\n\n")
	fmt.Fprintf(&b, "Research Query: %s\n\n", query)

	b.WriteString("## Search Results\n")
	for i, r := range results {
		fmt.Fprintf(&b, "%d. **%s** (%s)\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet)
	}

	if len(facts.Facts) > 0 {
		b.WriteString("## Knowledge Graph Facts\n")
		for _, f := range facts.Facts {
			fmt.Fprintf(&b, "- %s: %s\n", f.Name, f.FactText)
		}
		b.WriteString("\n")
	}

	b.WriteString("Provide a well-structured markdown summary with sections, key findings, and connections between topics. Include code examples where relevant.")
	return b.String()
}
