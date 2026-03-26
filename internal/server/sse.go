package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/urmzd/zoro/internal/models"
)

func writeSSE(c echo.Context, rx <-chan models.SSEEvent) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		log.Println("[sse] warning: ResponseWriter does not support Flush, events may be buffered")
	}

	for evt := range rx {
		data, err := json.Marshal(evt)
		if err != nil {
			continue
		}
		if _, err := fmt.Fprintf(c.Response(), "data: %s\n\n", data); err != nil {
			log.Printf("[sse] write error (client likely disconnected): %v", err)
			return nil
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
	return nil
}
