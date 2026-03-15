package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/urmzd/kgdk/kgtypes"
)

func (s *Server) SearchKnowledge(c echo.Context) error {
	q := c.QueryParam("q")
	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	log.Printf("[http] SearchKnowledge query=%q limit=%d", q, limit)
	resp, err := s.graph.SearchFacts(c.Request().Context(), q, kgtypes.WithLimit(limit))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetKnowledgeGraph(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := 300
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	log.Printf("[http] GetKnowledgeGraph limit=%d", limit)
	graph, err := s.graph.GetGraph(c.Request().Context(), int64(limit))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, graph)
}

func (s *Server) GetNodeDetail(c echo.Context) error {
	id := c.Param("id")
	depthStr := c.QueryParam("depth")
	depth := 1
	if depthStr != "" {
		if v, err := strconv.Atoi(depthStr); err == nil && v > 0 {
			depth = v
		}
	}
	log.Printf("[http] GetNodeDetail id=%s depth=%d", id, depth)
	detail, err := s.graph.GetNode(c.Request().Context(), id, depth)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, detail)
}
