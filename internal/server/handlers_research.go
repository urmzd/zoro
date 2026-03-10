package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) StartResearch(c echo.Context) error {
	var body struct {
		Query string `json:"query"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	log.Printf("[http] StartResearch query=%q", body.Query)

	session := s.orchestrator.CreateSession(body.Query)
	rx := s.orchestrator.Subscribe(session.ID)
	if rx == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("session %s not found", session.ID))
	}

	go s.orchestrator.Run(c.Request().Context(), session.ID)

	return writeSSE(c, rx)
}
