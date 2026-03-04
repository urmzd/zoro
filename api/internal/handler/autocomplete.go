package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/urmzd/zoro/api/internal/service"
)

type Autocomplete struct {
	ollama   *service.OllamaClient
	registry *service.ModelRegistry
}

func NewAutocomplete(ollama *service.OllamaClient, registry *service.ModelRegistry) *Autocomplete {
	return &Autocomplete{ollama: ollama, registry: registry}
}

type AutocompleteResponse struct {
	Suggestions []string `json:"suggestions"`
}

// Suggest returns autocomplete suggestions for a partial query.
func (h *Autocomplete) Suggest(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AutocompleteResponse{Suggestions: []string{}})
		return
	}

	prompt := `Given the partial query below, suggest 3 to 5 complete search queries the user might intend. Return ONLY a JSON array of strings, no extra text.

Partial query: ` + q

	raw, err := h.ollama.GenerateWithModel(r.Context(), prompt, h.registry.Model(service.TierFast))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AutocompleteResponse{Suggestions: []string{}})
		return
	}

	suggestions := parseJSONArray(raw)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AutocompleteResponse{Suggestions: suggestions})
}

func parseJSONArray(raw string) []string {
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end <= start {
		return []string{}
	}

	var arr []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &arr); err != nil {
		return []string{}
	}
	return arr
}
