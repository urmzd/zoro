package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	kgtypes "github.com/urmzd/saige/knowledge/types"
	"github.com/urmzd/zoro/internal/agent"
	"github.com/urmzd/zoro/internal/graph"
	"github.com/urmzd/zoro/internal/orchestrator"
	"github.com/urmzd/zoro/internal/searcher"
)

func NewServer(ag *agent.Agent, orch *orchestrator.Orchestrator, g kgtypes.Graph, s *searcher.Searcher) *gomcp.Server {
	srv := gomcp.NewServer(&gomcp.Implementation{Name: "zoro", Version: "0.1.0"}, nil)

	srv.AddTool(chatTool(), chatHandler(ag))
	srv.AddTool(researchTool(), researchHandler(orch))
	srv.AddTool(webSearchTool(), webSearchHandler(s))
	srv.AddTool(searchKnowledgeTool(), searchKnowledgeHandler(g))
	srv.AddTool(storeKnowledgeTool(), storeKnowledgeHandler(g))
	srv.AddTool(getKnowledgeGraphTool(), getKnowledgeGraphHandler(g))

	return srv
}

func textResult(text string) *gomcp.CallToolResult {
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: text}},
	}
}

func errorResult(msg string) *gomcp.CallToolResult {
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: msg}},
		IsError: true,
	}
}

func getArgs(req *gomcp.CallToolRequest) map[string]any {
	var args map[string]any
	if req.Params.Arguments != nil {
		_ = json.Unmarshal(req.Params.Arguments, &args)
	}
	if args == nil {
		args = make(map[string]any)
	}
	return args
}

// schema builds a JSON Schema object for tool input.
func schema(properties map[string]any, required ...string) json.RawMessage {
	s := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	b, _ := json.Marshal(s)
	return b
}

func prop(typ, desc string) map[string]any {
	return map[string]any{"type": typ, "description": desc}
}

// ── Tool definitions ───────────────────────────────────────────────

func chatTool() *gomcp.Tool {
	return &gomcp.Tool{
		Name:        "chat",
		Description: "Chat with Zoro, an AI research assistant. Searches the web and knowledge graph to answer questions.",
		InputSchema: schema(map[string]any{
			"message":    prop("string", "The message to send"),
			"session_id": prop("string", "Optional session ID to continue an existing conversation"),
		}, "message"),
	}
}

func researchTool() *gomcp.Tool {
	return &gomcp.Tool{
		Name:        "research",
		Description: "Run a deep research pipeline: web search, knowledge graph ingestion, and LLM synthesis.",
		InputSchema: schema(map[string]any{
			"query": prop("string", "The research query"),
		}, "query"),
	}
}

func webSearchTool() *gomcp.Tool {
	return &gomcp.Tool{
		Name:        "web_search",
		Description: "Search the web for current information on a topic.",
		InputSchema: schema(map[string]any{
			"query": prop("string", "The search query"),
		}, "query"),
	}
}

func searchKnowledgeTool() *gomcp.Tool {
	return &gomcp.Tool{
		Name:        "search_knowledge",
		Description: "Search the knowledge graph for previously stored facts and entities.",
		InputSchema: schema(map[string]any{
			"query": prop("string", "The search query"),
		}, "query"),
	}
}

func storeKnowledgeTool() *gomcp.Tool {
	return &gomcp.Tool{
		Name:        "store_knowledge",
		Description: "Store information into the knowledge graph by extracting entities and relationships from text.",
		InputSchema: schema(map[string]any{
			"text":   prop("string", "The text content to extract knowledge from"),
			"source": prop("string", "Description of the source of this information"),
		}, "text", "source"),
	}
}

func getKnowledgeGraphTool() *gomcp.Tool {
	return &gomcp.Tool{
		Name:        "get_knowledge_graph",
		Description: "Get the knowledge graph structure: all entities (nodes) and their relationships (edges). Returns a structured view of all stored knowledge that shows how concepts are connected. Use format=dot to get Graphviz DOT output for SVG rendering.",
		InputSchema: schema(map[string]any{
			"format":  prop("string", "Output format: text (default, human/AI-readable), dot (Graphviz DOT for SVG), json (structured data)"),
			"limit":   prop("number", "Max number of edges to return (default 100)"),
			"node_id": prop("string", "Optional: entity UUID to show only its neighborhood"),
			"depth":   prop("number", "Traversal depth when using node_id (default 2)"),
		}),
	}
}

// ── Handlers ───────────────────────────────────────────────────────

func chatHandler(ag *agent.Agent) gomcp.ToolHandler {
	return func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := getArgs(req)
		message, _ := args["message"].(string)
		if message == "" {
			return errorResult("message is required"), nil
		}
		sessionID, _ := args["session_id"].(string)

		response, returnedID, err := ag.InvokeSync(ctx, sessionID, message)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := fmt.Sprintf("%s\n\n---\nsession_id: %s", response, returnedID)
		return textResult(result), nil
	}
}

func researchHandler(orch *orchestrator.Orchestrator) gomcp.ToolHandler {
	return func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := getArgs(req)
		query, _ := args["query"].(string)
		if query == "" {
			return errorResult("query is required"), nil
		}

		summary, err := orch.RunSync(ctx, query)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(summary), nil
	}
}

func webSearchHandler(s *searcher.Searcher) gomcp.ToolHandler {
	return func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := getArgs(req)
		query, _ := args["query"].(string)
		if query == "" {
			return errorResult("query is required"), nil
		}

		results, err := s.Search(ctx, query)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		out, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(string(out)), nil
	}
}

func searchKnowledgeHandler(g kgtypes.Graph) gomcp.ToolHandler {
	return func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := getArgs(req)
		query, _ := args["query"].(string)
		if query == "" {
			return errorResult("query is required"), nil
		}

		resp, err := g.SearchFacts(ctx, query, kgtypes.WithLimit(10))
		if err != nil {
			return errorResult(err.Error()), nil
		}

		var b strings.Builder
		for _, f := range resp.Facts {
			fmt.Fprintf(&b, "- %s → %s: %s\n", f.SourceNode.Name, f.TargetNode.Name, f.FactText)
		}

		if b.Len() == 0 {
			return textResult("No relevant knowledge found."), nil
		}
		return textResult(b.String()), nil
	}
}

func getKnowledgeGraphHandler(g kgtypes.Graph) gomcp.ToolHandler {
	return func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := getArgs(req)
		format, _ := args["format"].(string)
		if format == "" {
			format = "text"
		}

		limit := int64(100)
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = int64(l)
		}

		depth := 2
		if d, ok := args["depth"].(float64); ok && d > 0 {
			depth = int(d)
		}

		var data *kgtypes.GraphData

		if nodeID, _ := args["node_id"].(string); nodeID != "" {
			detail, err := g.GetNode(ctx, nodeID, depth)
			if err != nil {
				return errorResult(err.Error()), nil
			}
			nodes := make([]kgtypes.GraphNode, 0, 1+len(detail.Neighbors))
			nodes = append(nodes, detail.Node)
			nodes = append(nodes, detail.Neighbors...)
			data = &kgtypes.GraphData{Nodes: nodes, Edges: detail.Edges}
		} else {
			var err error
			data, err = g.GetGraph(ctx, limit)
			if err != nil {
				return errorResult(err.Error()), nil
			}
		}

		var output string
		switch format {
		case "dot":
			output = graph.ToDOT(data)
		case "json":
			out, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return errorResult(err.Error()), nil
			}
			output = string(out)
		default:
			output = graph.ToText(data)
		}

		if output == "" {
			return textResult("Knowledge graph is empty."), nil
		}
		return textResult(output), nil
	}
}

func storeKnowledgeHandler(g kgtypes.Graph) gomcp.ToolHandler {
	return func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := getArgs(req)
		text, _ := args["text"].(string)
		if text == "" {
			return errorResult("text is required"), nil
		}
		source, _ := args["source"].(string)
		if source == "" {
			return errorResult("source is required"), nil
		}

		input := &kgtypes.EpisodeInput{
			Name:   source,
			Body:   text,
			Source: source,
		}

		resp, err := g.IngestEpisode(ctx, input)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("Stored %d entities and %d relations.", len(resp.EntityNodes), len(resp.EpisodicEdges))), nil
	}
}
