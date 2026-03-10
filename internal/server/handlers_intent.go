package server

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/urmzd/zoro/internal/models"
)

func (s *Server) ClassifyIntent(c echo.Context) error {
	var body struct {
		Query string `json:"query"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	log.Printf("[http] ClassifyIntent query=%q", body.Query)

	action, err := s.agent.ClassifyIntent(c.Request().Context(), body.Query)
	if err != nil {
		log.Printf("[http] ClassifyIntent error: %v", err)
		action = "chat"
	}
	return c.JSON(http.StatusOK, &models.IntentResponse{Action: action, Query: body.Query})
}

func (s *Server) GetAutocomplete(c echo.Context) error {
	q := c.QueryParam("q")
	log.Printf("[http] GetAutocomplete query=%q", q)
	suggestions := s.agent.Autocomplete(c.Request().Context(), q)
	return c.JSON(http.StatusOK, &models.AutocompleteResponse{Suggestions: suggestions})
}
