package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	kgtypes "github.com/urmzd/saige/knowledge/types"
	"github.com/urmzd/zoro/internal/searcher"
)

// mockGraph implements kgtypes.Graph for testing.
type mockGraph struct {
	searchResult *kgtypes.SearchFactsResult
	ingestResult *kgtypes.IngestResult
	graphData    *kgtypes.GraphData
	nodeDetail   *kgtypes.NodeDetail
	searchErr    error
	ingestErr    error
	graphErr     error
	nodeErr      error
}

func (m *mockGraph) ApplyOntology(ctx context.Context, ont *kgtypes.Ontology) error { return nil }
func (m *mockGraph) IngestEpisode(ctx context.Context, input *kgtypes.EpisodeInput) (*kgtypes.IngestResult, error) {
	if m.ingestErr != nil {
		return nil, m.ingestErr
	}
	if m.ingestResult != nil {
		return m.ingestResult, nil
	}
	return &kgtypes.IngestResult{}, nil
}
func (m *mockGraph) GetEntity(ctx context.Context, id string) (*kgtypes.Entity, error) {
	return &kgtypes.Entity{UUID: id}, nil
}
func (m *mockGraph) SearchFacts(ctx context.Context, query string, opts ...kgtypes.SearchOption) (*kgtypes.SearchFactsResult, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if m.searchResult != nil {
		return m.searchResult, nil
	}
	return &kgtypes.SearchFactsResult{}, nil
}
func (m *mockGraph) GetGraph(ctx context.Context, limit int64) (*kgtypes.GraphData, error) {
	if m.graphErr != nil {
		return nil, m.graphErr
	}
	if m.graphData != nil {
		return m.graphData, nil
	}
	return &kgtypes.GraphData{}, nil
}
func (m *mockGraph) GetNode(ctx context.Context, id string, depth int) (*kgtypes.NodeDetail, error) {
	if m.nodeErr != nil {
		return nil, m.nodeErr
	}
	if m.nodeDetail != nil {
		return m.nodeDetail, nil
	}
	return &kgtypes.NodeDetail{Node: kgtypes.GraphNode{ID: id, Name: "test"}}, nil
}
func (m *mockGraph) GetFactProvenance(ctx context.Context, factUUID string) ([]kgtypes.Episode, error) {
	return nil, nil
}
func (m *mockGraph) Close(ctx context.Context) error { return nil }

// helper to build a CallToolRequest with JSON arguments
func makeReq(args map[string]any) *gomcp.CallToolRequest {
	raw, _ := json.Marshal(args)
	return &gomcp.CallToolRequest{
		Params: &gomcp.CallToolParamsRaw{
			Arguments: raw,
		},
	}
}

func getText(t *testing.T, result *gomcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("empty content")
	}
	b, err := result.Content[0].MarshalJSON()
	if err != nil {
		t.Fatalf("marshal content: %v", err)
	}
	var c struct {
		Text string `json:"text"`
	}
	_ = json.Unmarshal(b, &c)
	return c.Text
}

// ── webSearchHandler tests ─────────────────────────────────────────

func TestWebSearchHandler_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]string{
				{"title": "Test", "url": "https://example.com", "content": "Snippet"},
			},
		})
	}))
	defer srv.Close()

	s := searcher.New(srv.URL)
	handler := webSearchHandler(s)

	result, err := handler(context.Background(), makeReq(map[string]any{"query": "test"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("got error result: %s", getText(t, result))
	}
}

func TestWebSearchHandler_EmptyQuery(t *testing.T) {
	s := searcher.New("http://unused")
	handler := webSearchHandler(s)

	result, err := handler(context.Background(), makeReq(map[string]any{"query": ""}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for empty query")
	}
}

// ── searchKnowledgeHandler tests ───────────────────────────────────

func TestSearchKnowledgeHandler_WithFacts(t *testing.T) {
	mg := &mockGraph{
		searchResult: &kgtypes.SearchFactsResult{
			Facts: []kgtypes.Fact{
				{
					SourceNode: kgtypes.Entity{Name: "A"},
					TargetNode: kgtypes.Entity{Name: "B"},
					FactText:   "A relates to B",
				},
			},
		},
	}
	handler := searchKnowledgeHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{"query": "test"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	if !strings.Contains(text, "A relates to B") {
		t.Errorf("expected fact text in result: %q", text)
	}
}

func TestSearchKnowledgeHandler_NoFacts(t *testing.T) {
	mg := &mockGraph{searchResult: &kgtypes.SearchFactsResult{}}
	handler := searchKnowledgeHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{"query": "nothing"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	if text != "No relevant knowledge found." {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestSearchKnowledgeHandler_EmptyQuery(t *testing.T) {
	handler := searchKnowledgeHandler(&mockGraph{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"query": ""}))
	if !result.IsError {
		t.Error("expected error for empty query")
	}
}

// ── storeKnowledgeHandler tests ────────────────────────────────────

func TestStoreKnowledgeHandler_Success(t *testing.T) {
	mg := &mockGraph{
		ingestResult: &kgtypes.IngestResult{
			EntityNodes:   []kgtypes.Entity{{Name: "X"}},
			EpisodicEdges: []kgtypes.Relation{},
		},
	}
	handler := storeKnowledgeHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{
		"text":   "some knowledge",
		"source": "test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	if !strings.Contains(text, "1 entities") {
		t.Errorf("unexpected result: %q", text)
	}
}

func TestStoreKnowledgeHandler_EmptyText(t *testing.T) {
	handler := storeKnowledgeHandler(&mockGraph{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"text": "", "source": "x"}))
	if !result.IsError {
		t.Error("expected error for empty text")
	}
}

func TestStoreKnowledgeHandler_EmptySource(t *testing.T) {
	handler := storeKnowledgeHandler(&mockGraph{})
	result, _ := handler(context.Background(), makeReq(map[string]any{"text": "hello", "source": ""}))
	if !result.IsError {
		t.Error("expected error for empty source")
	}
}

// ── getKnowledgeGraphHandler tests ─────────────────────────────────

func TestGetKnowledgeGraphHandler_TextFormat(t *testing.T) {
	mg := &mockGraph{
		graphData: &kgtypes.GraphData{
			Nodes: []kgtypes.GraphNode{{ID: "1", Name: "Test", Type: "Entity"}},
		},
	}
	handler := getKnowledgeGraphHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	if !strings.Contains(text, "Test") {
		t.Errorf("expected entity name: %q", text)
	}
}

func TestGetKnowledgeGraphHandler_DOTFormat(t *testing.T) {
	mg := &mockGraph{
		graphData: &kgtypes.GraphData{
			Nodes: []kgtypes.GraphNode{{ID: "1", Name: "Test"}},
		},
	}
	handler := getKnowledgeGraphHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{"format": "dot"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	if !strings.Contains(text, "digraph") {
		t.Errorf("expected DOT output: %q", text)
	}
}

func TestGetKnowledgeGraphHandler_JSONFormat(t *testing.T) {
	mg := &mockGraph{
		graphData: &kgtypes.GraphData{
			Nodes: []kgtypes.GraphNode{{ID: "1", Name: "Test"}},
		},
	}
	handler := getKnowledgeGraphHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{"format": "json"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	var data kgtypes.GraphData
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestGetKnowledgeGraphHandler_WithNodeID(t *testing.T) {
	mg := &mockGraph{
		nodeDetail: &kgtypes.NodeDetail{
			Node:      kgtypes.GraphNode{ID: "n1", Name: "Center"},
			Neighbors: []kgtypes.GraphNode{{ID: "n2", Name: "Neighbor"}},
			Edges:     []kgtypes.GraphEdge{{Source: "n1", Target: "n2", Type: "links"}},
		},
	}
	handler := getKnowledgeGraphHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{"node_id": "n1"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	if !strings.Contains(text, "Center") {
		t.Errorf("expected center node name: %q", text)
	}
}

func TestGetKnowledgeGraphHandler_Empty(t *testing.T) {
	mg := &mockGraph{graphData: &kgtypes.GraphData{}}
	handler := getKnowledgeGraphHandler(mg)

	result, err := handler(context.Background(), makeReq(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getText(t, result)
	if text == "" {
		t.Error("expected non-empty text")
	}
}

// ── NewServer tests ────────────────────────────────────────────────

func TestNewServer_RegistersTools(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{}})
	}))
	defer srv.Close()

	mg := &mockGraph{}
	s := searcher.New(srv.URL)

	mcpSrv := NewServer(nil, nil, mg, s)
	if mcpSrv == nil {
		t.Fatal("NewServer returned nil")
	}
}
