package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) GetStatus(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	status := ServiceStatus{
		SurrealDB: s.graph != nil,
		SearXNG:   checkHealth(ctx, s.searxngURL),
		Ollama:    checkHealth(ctx, s.ollamaHost+"/api/tags"),
	}
	return c.JSON(http.StatusOK, status)
}

func checkHealth(ctx context.Context, url string) bool {
	if url == "" {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}
