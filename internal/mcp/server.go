package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	kgtypes "github.com/urmzd/saige/knowledge/types"
	"github.com/urmzd/zoro/internal/agent"
	"github.com/urmzd/zoro/internal/graph"
	"github.com/urmzd/zoro/internal/orchestrator"
	"github.com/urmzd/zoro/internal/searcher"
)

func NewServer(ag *agent.Agent, orch *orchestrator.Orchestrator, graph kgtypes.Graph, s *searcher.Searcher) *server.MCPServer {
	srv := server.NewMCPServer("zoro", "0.1.0", server.WithToolCapabilities(false))

	srv.AddTools(
		server.ServerTool{
			Tool: mcp.NewTool("chat",
				mcp.WithDescription("Chat with Zoro, an AI research assistant. Searches the web and knowledge graph to answer questions."),
				mcp.WithString("message", mcp.Required(), mcp.Description("The message to send")),
				mcp.WithString("session_id", mcp.Description("Optional session ID to continue an existing conversation")),
			),
			Handler: chatHandler(ag),
		},
		server.ServerTool{
			Tool: mcp.NewTool("research",
				mcp.WithDescription("Run a deep research pipeline: web search, knowledge graph ingestion, and LLM synthesis."),
				mcp.WithString("query", mcp.Required(), mcp.Description("The research query")),
			),
			Handler: researchHandler(orch),
		},
		server.ServerTool{
			Tool: mcp.NewTool("web_search",
				mcp.WithDescription("Search the web for current information on a topic."),
				mcp.WithString("query", mcp.Required(), mcp.Description("The search query")),
			),
			Handler: webSearchHandler(s),
		},
		server.ServerTool{
			Tool: mcp.NewTool("search_knowledge",
				mcp.WithDescription("Search the knowledge graph for previously stored facts and entities."),
				mcp.WithString("query", mcp.Required(), mcp.Description("The search query")),
			),
			Handler: searchKnowledgeHandler(graph),
		},
		server.ServerTool{
			Tool: mcp.NewTool("store_knowledge",
				mcp.WithDescription("Store information into the knowledge graph by extracting entities and relationships from text."),
				mcp.WithString("text", mcp.Required(), mcp.Description("The text content to extract knowledge from")),
				mcp.WithString("source", mcp.Required(), mcp.Description("Description of the source of this information")),
			),
			Handler: storeKnowledgeHandler(graph),
		},
		server.ServerTool{
			Tool: mcp.NewTool("get_knowledge_graph",
				mcp.WithDescription("Get the knowledge graph structure: all entities (nodes) and their relationships (edges). Returns a structured view of all stored knowledge that shows how concepts are connected. Use format=dot to get Graphviz DOT output for SVG rendering."),
				mcp.WithString("format", mcp.Description("Output format: text (default, human/AI-readable), dot (Graphviz DOT for SVG), json (structured data)")),
				mcp.WithNumber("limit", mcp.Description("Max number of edges to return (default 100)")),
				mcp.WithString("node_id", mcp.Description("Optional: entity UUID to show only its neighborhood")),
				mcp.WithNumber("depth", mcp.Description("Traversal depth when using node_id (default 2)")),
			),
			Handler: getKnowledgeGraphHandler(graph),
		},
	)

	return srv
}

func chatHandler(ag *agent.Agent) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		message, _ := args["message"].(string)
		if message == "" {
			return mcp.NewToolResultError("message is required"), nil
		}
		sessionID, _ := args["session_id"].(string)

		response, returnedID, err := ag.InvokeSync(ctx, sessionID, message)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := fmt.Sprintf("%s\n\n---\nsession_id: %s", response, returnedID)
		return mcp.NewToolResultText(result), nil
	}
}

func researchHandler(orch *orchestrator.Orchestrator) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		query, _ := args["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		summary, err := orch.RunSync(ctx, query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(summary), nil
	}
}

func webSearchHandler(s *searcher.Searcher) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		query, _ := args["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		results, err := s.Search(ctx, query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		out, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}
}

func searchKnowledgeHandler(graph kgtypes.Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		query, _ := args["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		resp, err := graph.SearchFacts(ctx, query, kgtypes.WithLimit(10))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var b strings.Builder
		for _, f := range resp.Facts {
			fmt.Fprintf(&b, "- %s → %s: %s\n", f.SourceNode.Name, f.TargetNode.Name, f.FactText)
		}

		if b.Len() == 0 {
			return mcp.NewToolResultText("No relevant knowledge found."), nil
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func getKnowledgeGraphHandler(g kgtypes.Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
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
				return mcp.NewToolResultError(err.Error()), nil
			}
			nodes := make([]kgtypes.GraphNode, 0, 1+len(detail.Neighbors))
			nodes = append(nodes, detail.Node)
			nodes = append(nodes, detail.Neighbors...)
			data = &kgtypes.GraphData{Nodes: nodes, Edges: detail.Edges}
		} else {
			var err error
			data, err = g.GetGraph(ctx, limit)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		}

		var output string
		switch format {
		case "dot":
			output = graph.ToDOT(data)
		case "json":
			out, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			output = string(out)
		default:
			output = graph.ToText(data)
		}

		if output == "" {
			return mcp.NewToolResultText("Knowledge graph is empty."), nil
		}
		return mcp.NewToolResultText(output), nil
	}
}

func storeKnowledgeHandler(graph kgtypes.Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		text, _ := args["text"].(string)
		if text == "" {
			return mcp.NewToolResultError("text is required"), nil
		}
		source, _ := args["source"].(string)
		if source == "" {
			return mcp.NewToolResultError("source is required"), nil
		}

		input := &kgtypes.EpisodeInput{
			Name:   source,
			Body:   text,
			Source: source,
		}

		resp, err := graph.IngestEpisode(ctx, input)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Stored %d entities and %d relations.", len(resp.EntityNodes), len(resp.EpisodicEdges))), nil
	}
}
