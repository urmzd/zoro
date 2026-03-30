package graph

import (
	"strings"
	"testing"

	kgtypes "github.com/urmzd/saige/knowledge/types"
)

func sampleData() *kgtypes.GraphData {
	return &kgtypes.GraphData{
		Nodes: []kgtypes.GraphNode{
			{ID: "1", Name: "Go", Type: "Language"},
			{ID: "2", Name: "Rust", Type: "Language", Summary: "Systems language"},
			{ID: "3", Name: "LLVM", Type: "Compiler"},
		},
		Edges: []kgtypes.GraphEdge{
			{ID: "e1", Source: "2", Target: "3", Type: "uses", Fact: "Rust uses LLVM backend"},
		},
	}
}

func TestToDOT_Structure(t *testing.T) {
	dot := ToDOT(sampleData())

	if !strings.HasPrefix(dot, "digraph knowledge {") {
		t.Error("DOT output should start with digraph declaration")
	}
	if !strings.HasSuffix(dot, "}\n") {
		t.Error("DOT output should end with closing brace")
	}
	if !strings.Contains(dot, `"1"`) {
		t.Error("DOT should contain node ID 1")
	}
	if !strings.Contains(dot, `"2" -> "3"`) {
		t.Error("DOT should contain edge from 2 to 3")
	}
	if !strings.Contains(dot, `label="uses"`) {
		t.Error("DOT should contain edge label")
	}
	if !strings.Contains(dot, `Go\\n[Language]`) {
		t.Error("DOT node label should include type")
	}
}

func TestToDOT_Empty(t *testing.T) {
	dot := ToDOT(&kgtypes.GraphData{})
	if !strings.Contains(dot, "digraph knowledge {") {
		t.Error("empty graph should still produce valid DOT")
	}
}

func TestToText_Structure(t *testing.T) {
	text := ToText(sampleData())

	if !strings.Contains(text, "3 nodes, 1 edges") {
		t.Error("text should contain node/edge counts")
	}
	if !strings.Contains(text, "## Entities") {
		t.Error("text should contain Entities section")
	}
	if !strings.Contains(text, "## Relations") {
		t.Error("text should contain Relations section")
	}
	if !strings.Contains(text, "Go (Language)") {
		t.Error("text should show entity with type")
	}
	if !strings.Contains(text, "Rust (Language): Systems language") {
		t.Error("text should show entity with summary")
	}
	if !strings.Contains(text, "Rust -[uses]-> LLVM: Rust uses LLVM backend") {
		t.Error("text should show relation with names resolved and fact")
	}
}

func TestToText_Empty(t *testing.T) {
	text := ToText(&kgtypes.GraphData{})
	if !strings.Contains(text, "0 nodes, 0 edges") {
		t.Error("empty graph text should show zero counts")
	}
}

func TestToText_UnknownNodeID(t *testing.T) {
	data := &kgtypes.GraphData{
		Edges: []kgtypes.GraphEdge{
			{Source: "unknown-src", Target: "unknown-tgt", Type: "relates"},
		},
	}
	text := ToText(data)
	if !strings.Contains(text, "unknown-src -[relates]-> unknown-tgt") {
		t.Error("should fall back to raw IDs for unknown nodes")
	}
}

func TestEscDOT(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"simple", "simple"},
		{`has "quotes"`, `has \"quotes\"`},
		{`back\slash`, `back\\slash`},
		{`both "and" \\ here`, `both \"and\" \\\\ here`},
	}
	for _, tt := range tests {
		got := escDOT(tt.in)
		if got != tt.want {
			t.Errorf("escDOT(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
