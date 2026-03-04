package handler

import (
	"encoding/json"
	"net/http"

	"github.com/urmzd/zoro/api/internal/service"
)

type Intent struct {
	agent *service.Agent
}

func NewIntent(agent *service.Agent) *Intent {
	return &Intent{agent: agent}
}

type IntentRequest struct {
	Query string `json:"query"`
}

type IntentResponse struct {
	Action string `json:"action"` // "chat" or "knowledge_search"
	Query  string `json:"query"`
}

// Classify determines the user's intent and returns a structured action.
func (h *Intent) Classify(w http.ResponseWriter, r *http.Request) {
	var req IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		http.Error(w, `{"error":"query is required"}`, http.StatusBadRequest)
		return
	}

	action, err := h.agent.ClassifyIntent(r.Context(), req.Query)
	if err != nil {
		// Default to chat on error
		action = "chat"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IntentResponse{
		Action: action,
		Query:  req.Query,
	})
}
