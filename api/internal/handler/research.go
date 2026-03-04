package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/urmzd/zoro/api/internal/model"
	"github.com/urmzd/zoro/api/internal/service"
)

type Research struct {
	orchestrator *service.Orchestrator
}

func NewResearch(orchestrator *service.Orchestrator) *Research {
	return &Research{orchestrator: orchestrator}
}

// StartResearchRequest is the request body for starting a research session.
type StartResearchRequest struct {
	Query string `json:"query" example:"What is quantum computing?"`
}

// Start creates a new research session and begins processing.
//
//	@Summary		Start research session
//	@Description	Creates a new research session for the given query
//	@Tags			research
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StartResearchRequest	true	"Research query"
//	@Success		200		{object}	map[string]string		"Session ID"
//	@Failure		400		{object}	map[string]string		"Bad request"
//	@Router			/api/research [post]
func (h *Research) Start(w http.ResponseWriter, r *http.Request) {
	var req StartResearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		http.Error(w, `{"error":"query is required"}`, http.StatusBadRequest)
		return
	}

	session := h.orchestrator.CreateSession(req.Query)

	go h.orchestrator.Run(context.Background(), session.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": session.ID})
}

// Get returns a research session by ID.
//
//	@Summary		Get research session
//	@Description	Returns the research session with the given ID
//	@Tags			research
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	model.ResearchSession
//	@Failure		404	{object}	map[string]string	"Session not found"
//	@Router			/api/research/{id} [get]
func (h *Research) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	session, ok := h.orchestrator.GetSession(id)
	if !ok {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// Stream returns a server-sent event stream of research progress.
//
//	@Summary		Stream research events
//	@Description	Returns an SSE stream of research progress events
//	@Tags			research
//	@Produce		text/event-stream
//	@Param			id	path	string	true	"Session ID"
//	@Success		200	{object}	model.SSEEvent
//	@Failure		404	{object}	map[string]string	"Session not found"
//	@Router			/api/research/{id}/stream [get]
func (h *Research) Stream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ch, ok := h.orchestrator.Subscribe(id)
	if !ok {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
			flusher.Flush()
		}
	}
}
