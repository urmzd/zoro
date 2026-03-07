package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/urmzd/zoro/api/internal/service"
)

type Chat struct {
	agent *service.Agent
}

func NewChat(agent *service.Agent) *Chat {
	return &Chat{agent: agent}
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

// ListSessions returns all chat sessions.
func (h *Chat) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.agent.ListSessions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// CreateSession creates a new chat session.
func (h *Chat) CreateSession(w http.ResponseWriter, r *http.Request) {
	session := h.agent.CreateSession()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// GetSession returns a chat session by ID.
func (h *Chat) GetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	session, ok := h.agent.GetSession(id)
	if !ok {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// SendMessage sends a user message and streams the agent response as SSE.
func (h *Chat) SendMessage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		http.Error(w, `{"error":"content is required"}`, http.StatusBadRequest)
		return
	}

	// Subscribe BEFORE sending to avoid race condition
	ch, ok := h.agent.Subscribe(id)
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

	// Run agent in background
	go h.agent.SendMessage(context.Background(), id, req.Content)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, open := <-ch:
			if !open {
				return
			}
			data, _ := json.Marshal(evt.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
			flusher.Flush()
		}
	}
}
