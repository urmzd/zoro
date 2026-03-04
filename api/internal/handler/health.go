package handler

import (
	"encoding/json"
	"net/http"
)

type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

// Check returns the health status of the API.
//
//	@Summary		Health check
//	@Description	Returns the health status of the API
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (h *Health) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
