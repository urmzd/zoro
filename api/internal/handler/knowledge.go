package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	_ "github.com/urmzd/zoro/api/internal/model"
	"github.com/urmzd/zoro/api/internal/service"
)

type Knowledge struct {
	knowledge service.KnowledgeStore
}

func NewKnowledge(knowledge service.KnowledgeStore) *Knowledge {
	return &Knowledge{knowledge: knowledge}
}

// Search searches the knowledge graph for facts matching the query.
//
//	@Summary		Search knowledge graph
//	@Description	Searches for facts in the knowledge graph
//	@Tags			knowledge
//	@Produce		json
//	@Param			q	query		string	true	"Search query"
//	@Success		200	{object}	model.SearchFactsResponse
//	@Failure		400	{object}	map[string]string	"Missing query"
//	@Failure		500	{object}	map[string]string	"Search failed"
//	@Router			/api/knowledge/search [get]
func (h *Knowledge) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, `{"error":"q parameter is required"}`, http.StatusBadRequest)
		return
	}

	results, err := h.knowledge.SearchFacts(r.Context(), q, "")
	if err != nil {
		http.Error(w, `{"error":"search failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Graph returns the knowledge graph data.
//
//	@Summary		Get knowledge graph
//	@Description	Returns nodes and edges from the knowledge graph
//	@Tags			knowledge
//	@Produce		json
//	@Param			limit	query		int		false	"Max number of nodes"	default(300)
//	@Success		200		{object}	model.GraphData
//	@Failure		500		{object}	map[string]string	"Fetch failed"
//	@Router			/api/knowledge/graph [get]
func (h *Knowledge) Graph(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 300
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	graph, err := h.knowledge.GetGraph(r.Context(), limit)
	if err != nil {
		http.Error(w, `{"error":"graph fetch failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}

// Node returns details for a specific node in the knowledge graph.
//
//	@Summary		Get node detail
//	@Description	Returns a node and its neighbors from the knowledge graph
//	@Tags			knowledge
//	@Produce		json
//	@Param			id		path		string	true	"Node ID"
//	@Param			depth	query		int		false	"Neighbor depth"	default(1)
//	@Success		200		{object}	model.NodeDetail
//	@Failure		500		{object}	map[string]string	"Node not found"
//	@Router			/api/knowledge/node/{id} [get]
func (h *Knowledge) Node(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	depthStr := r.URL.Query().Get("depth")
	depth := 1
	if depthStr != "" {
		if n, err := strconv.Atoi(depthStr); err == nil && n > 0 {
			depth = n
		}
	}

	node, err := h.knowledge.GetNode(r.Context(), id, depth)
	if err != nil {
		http.Error(w, `{"error":"node not found"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}
