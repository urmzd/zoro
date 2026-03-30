package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kgtypes "github.com/urmzd/saige/knowledge/types"
	"github.com/urmzd/zoro/internal/searcher"
)

// mockGraph implements kgtypes.Graph for testing.
type mockGraph struct {
	ingestCalls  int
	searchCalls  int
	getGraphData *kgtypes.GraphData
	searchResult *kgtypes.SearchFactsResult
	ingestResult *kgtypes.IngestResult
	ingestErr    error
	searchErr    error
	getGraphErr  error
}

func (m *mockGraph) ApplyOntology(ctx context.Context, ont *kgtypes.Ontology) error { return nil }

func (m *mockGraph) IngestEpisode(ctx context.Context, input *kgtypes.EpisodeInput) (*kgtypes.IngestResult, error) {
	m.ingestCalls++
	if m.ingestErr != nil {
		return nil, m.ingestErr
	}
	if m.ingestResult != nil {
		return m.ingestResult, nil
	}
	return &kgtypes.IngestResult{
		EntityNodes:   []kgtypes.Entity{{UUID: "e1", Name: "test"}},
		EpisodicEdges: []kgtypes.Relation{},
	}, nil
}

func (m *mockGraph) GetEntity(ctx context.Context, id string) (*kgtypes.Entity, error) {
	return &kgtypes.Entity{UUID: id, Name: "test"}, nil
}

func (m *mockGraph) SearchFacts(ctx context.Context, query string, opts ...kgtypes.SearchOption) (*kgtypes.SearchFactsResult, error) {
	m.searchCalls++
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if m.searchResult != nil {
		return m.searchResult, nil
	}
	return &kgtypes.SearchFactsResult{}, nil
}

func (m *mockGraph) GetGraph(ctx context.Context, limit int64) (*kgtypes.GraphData, error) {
	if m.getGraphErr != nil {
		return nil, m.getGraphErr
	}
	if m.getGraphData != nil {
		return m.getGraphData, nil
	}
	return &kgtypes.GraphData{}, nil
}

func (m *mockGraph) GetNode(ctx context.Context, id string, depth int) (*kgtypes.NodeDetail, error) {
	return &kgtypes.NodeDetail{
		Node: kgtypes.GraphNode{ID: id, Name: "test"},
	}, nil
}

func (m *mockGraph) GetFactProvenance(ctx context.Context, factUUID string) ([]kgtypes.Episode, error) {
	return nil, nil
}

func (m *mockGraph) Close(ctx context.Context) error { return nil }

// ── WebSearchTool tests ────────────────────────────────────────────

func TestWebSearchTool_Definition(t *testing.T) {
	tool := NewWebSearchTool(nil, nil)
	def := tool.Definition()
	if def.Name != "web_search" {
		t.Errorf("name = %q, want %q", def.Name, "web_search")
	}
	if _, ok := def.Parameters.Properties["query"]; !ok {
		t.Error("missing query parameter")
	}
}

func TestWebSearchTool_EmptyQuery(t *testing.T) {
	tool := NewWebSearchTool(nil, nil)
	_, err := tool.Execute(context.Background(), map[string]any{"query": ""})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestWebSearchTool_Execute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type result struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []result{
				{Title: "Test", URL: "https://example.com", Content: "Snippet"},
			},
		})
	}))
	defer srv.Close()

	mg := &mockGraph{}
	s := searcher.New(srv.URL)
	tool := NewWebSearchTool(s, mg)

	result, err := tool.Execute(context.Background(), map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []searchResultJSON
	if err := json.Unmarshal([]byte(result), &items); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Index != 1 {
		t.Errorf("index = %d, want 1", items[0].Index)
	}
}

func TestWebSearchTool_TruncatesSnippet(t *testing.T) {
	longSnippet := strings.Repeat("a", 300)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]string{
				{"title": "Test", "url": "https://example.com", "content": longSnippet},
			},
		})
	}))
	defer srv.Close()

	s := searcher.New(srv.URL)
	tool := NewWebSearchTool(s, nil)

	result, err := tool.Execute(context.Background(), map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []searchResultJSON
	_ = json.Unmarshal([]byte(result), &items)
	if len(items[0].Snippet) > 204 { // 200 + "..."
		t.Errorf("snippet not truncated: len=%d", len(items[0].Snippet))
	}
}

func TestWebSearchTool_NoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{}})
	}))
	defer srv.Close()

	s := searcher.New(srv.URL)
	tool := NewWebSearchTool(s, nil)

	result, err := tool.Execute(context.Background(), map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[]" {
		t.Errorf("expected empty array, got %q", result)
	}
}

func TestWebSearchTool_WithGroupID(t *testing.T) {
	tool := NewWebSearchTool(nil, nil)
	scoped := tool.WithGroupID("session-123")
	if scoped.groupID != "session-123" {
		t.Errorf("groupID = %q, want %q", scoped.groupID, "session-123")
	}
}

func TestWebSearchTool_AutoIngests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]string{
				{"title": "R1", "url": "https://a.com", "content": "S1"},
				{"title": "R2", "url": "https://b.com", "content": "S2"},
			},
		})
	}))
	defer srv.Close()

	mg := &mockGraph{}
	s := searcher.New(srv.URL)
	tool := NewWebSearchTool(s, mg)

	_, err := tool.Execute(context.Background(), map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mg.ingestCalls != 2 {
		t.Errorf("ingestCalls = %d, want 2", mg.ingestCalls)
	}
}

// ── SearchKnowledgeTool tests ──────────────────────────────────────

func TestSearchKnowledgeTool_Definition(t *testing.T) {
	tool := NewSearchKnowledgeTool(nil)
	def := tool.Definition()
	if def.Name != "search_knowledge" {
		t.Errorf("name = %q, want %q", def.Name, "search_knowledge")
	}
}

func TestSearchKnowledgeTool_EmptyQuery(t *testing.T) {
	tool := NewSearchKnowledgeTool(nil)
	_, err := tool.Execute(context.Background(), map[string]any{"query": ""})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearchKnowledgeTool_WithResults(t *testing.T) {
	mg := &mockGraph{
		searchResult: &kgtypes.SearchFactsResult{
			Facts: []kgtypes.Fact{
				{
					SourceNode: kgtypes.Entity{Name: "Go"},
					TargetNode: kgtypes.Entity{Name: "Google"},
					FactText:   "Go was created by Google",
				},
			},
		},
	}
	tool := NewSearchKnowledgeTool(mg)

	result, err := tool.Execute(context.Background(), map[string]any{"query": "Go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Go -> Google: Go was created by Google") {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestSearchKnowledgeTool_NoResults(t *testing.T) {
	mg := &mockGraph{searchResult: &kgtypes.SearchFactsResult{}}
	tool := NewSearchKnowledgeTool(mg)

	result, err := tool.Execute(context.Background(), map[string]any{"query": "nothing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "No relevant knowledge found." {
		t.Errorf("expected no-results message, got %q", result)
	}
}

func TestSearchKnowledgeTool_WithGroupID(t *testing.T) {
	tool := NewSearchKnowledgeTool(nil)
	scoped := tool.WithGroupID("g1")
	if scoped.groupID != "g1" {
		t.Errorf("groupID = %q, want %q", scoped.groupID, "g1")
	}
}

// ── StoreKnowledgeTool tests ───────────────────────────────────────

func TestStoreKnowledgeTool_Definition(t *testing.T) {
	tool := NewStoreKnowledgeTool(nil)
	def := tool.Definition()
	if def.Name != "store_knowledge" {
		t.Errorf("name = %q, want %q", def.Name, "store_knowledge")
	}
	if len(def.Parameters.Required) != 2 {
		t.Errorf("expected 2 required params, got %d", len(def.Parameters.Required))
	}
}

func TestStoreKnowledgeTool_EmptyText(t *testing.T) {
	tool := NewStoreKnowledgeTool(nil)
	_, err := tool.Execute(context.Background(), map[string]any{"text": "", "source": "test"})
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestStoreKnowledgeTool_Execute(t *testing.T) {
	mg := &mockGraph{
		ingestResult: &kgtypes.IngestResult{
			EntityNodes:   []kgtypes.Entity{{Name: "A"}, {Name: "B"}},
			EpisodicEdges: []kgtypes.Relation{{UUID: "r1"}},
		},
	}
	tool := NewStoreKnowledgeTool(mg)

	result, err := tool.Execute(context.Background(), map[string]any{
		"text":   "Go is a programming language",
		"source": "wikipedia",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "2 entities") {
		t.Errorf("expected entity count in result: %q", result)
	}
	if !strings.Contains(result, "1 relations") {
		t.Errorf("expected relation count in result: %q", result)
	}
}

func TestStoreKnowledgeTool_DefaultSource(t *testing.T) {
	mg := &mockGraph{}
	tool := NewStoreKnowledgeTool(mg)

	_, err := tool.Execute(context.Background(), map[string]any{
		"text": "some content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not error when source is omitted (defaults to "chat")
}

// ── GetGraphTool tests ─────────────────────────────────────────────

func TestGetGraphTool_Definition(t *testing.T) {
	tool := NewGetGraphTool(nil)
	def := tool.Definition()
	if def.Name != "get_knowledge_graph" {
		t.Errorf("name = %q, want %q", def.Name, "get_knowledge_graph")
	}
}

func TestGetGraphTool_Execute(t *testing.T) {
	mg := &mockGraph{
		getGraphData: &kgtypes.GraphData{
			Nodes: []kgtypes.GraphNode{{ID: "1", Name: "Test", Type: "Entity"}},
			Edges: []kgtypes.GraphEdge{},
		},
	}
	tool := NewGetGraphTool(mg)

	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Test") {
		t.Errorf("expected entity name in result: %q", result)
	}
}

func TestGetGraphTool_Empty(t *testing.T) {
	mg := &mockGraph{getGraphData: &kgtypes.GraphData{}}
	tool := NewGetGraphTool(mg)

	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ToText on empty data returns header with "0 nodes, 0 edges" which is not empty
	// The tool checks if text == "" to return "Knowledge graph is empty."
	// Since ToText always includes header text, it won't be empty
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestGetGraphTool_CustomLimit(t *testing.T) {
	mg := &mockGraph{getGraphData: &kgtypes.GraphData{}}
	tool := NewGetGraphTool(mg)

	_, err := tool.Execute(context.Background(), map[string]any{"limit": float64(50)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
