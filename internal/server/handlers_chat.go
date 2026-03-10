package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (s *Server) CreateSession(c echo.Context) error {
	log.Println("[http] CreateSession")
	session, err := s.agent.CreateSession()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, session)
}

func (s *Server) ListSessions(c echo.Context) error {
	log.Println("[http] ListSessions")
	sessions, err := s.agent.ListSessions()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, sessions)
}

func (s *Server) SearchSessions(c echo.Context) error {
	q := c.QueryParam("q")
	log.Printf("[http] SearchSessions q=%q", q)

	sessions, err := s.agent.ListSessions()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if q == "" {
		return c.JSON(http.StatusOK, sessions)
	}
	query := strings.ToLower(q)
	filtered := sessions[:0]
	for _, sess := range sessions {
		if strings.Contains(strings.ToLower(sess.Preview), query) {
			filtered = append(filtered, sess)
		}
	}
	return c.JSON(http.StatusOK, filtered)
}

func (s *Server) GetSession(c echo.Context) error {
	id := c.Param("id")
	log.Printf("[http] GetSession id=%s", id)
	session, err := s.agent.GetSession(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, session)
}

func (s *Server) SendMessage(c echo.Context) error {
	id := c.Param("id")
	var body struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	log.Printf("[http] SendMessage id=%s content_len=%d", id, len(body.Content))

	rx := s.agent.Subscribe(id)
	go s.agent.SendMessage(c.Request().Context(), id, body.Content)

	return writeSSE(c, rx)
}
