package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/urmzd/zoro/internal/models"
)

func writeSSE(c echo.Context, rx <-chan models.SSEEvent) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	flusher, _ := c.Response().Writer.(http.Flusher)

	for evt := range rx {
		data, err := json.Marshal(evt)
		if err != nil {
			continue
		}
		fmt.Fprintf(c.Response(), "data: %s\n\n", data)
		if flusher != nil {
			flusher.Flush()
		}
	}
	return nil
}
