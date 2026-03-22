package server

import (
	"github.com/labstack/echo/v4"
	"github.com/urmzd/adk/provider/ollama"
	"github.com/urmzd/kgdk/kgtypes"
	"github.com/urmzd/zoro/internal/agent"
	"github.com/urmzd/zoro/internal/orchestrator"
)

// ServiceStatus tracks readiness of external services.
type ServiceStatus struct {
	SurrealDB bool `json:"surrealdb"`
	SearXNG   bool `json:"searxng"`
	Ollama    bool `json:"ollama"`
}

type Server struct {
	agent        *agent.Agent
	orchestrator *orchestrator.Orchestrator
	graph        kgtypes.Graph
	adapter      *ollama.Adapter
	services     *ServiceStatus
}

func New(a *agent.Agent, o *orchestrator.Orchestrator, g kgtypes.Graph, ad *ollama.Adapter) *Server {
	return &Server{
		agent:        a,
		orchestrator: o,
		graph:        g,
		adapter:      ad,
		services:     &ServiceStatus{},
	}
}

// SetServiceStatus updates the readiness state of services.
func (s *Server) SetServiceStatus(status ServiceStatus) {
	*s.services = status
}

func (s *Server) Setup() *echo.Echo {
	e := echo.New()
	registerMiddleware(e)

	api := e.Group("/api")

	sessions := api.Group("/sessions")
	sessions.POST("", s.CreateSession)
	sessions.GET("", s.ListSessions)
	sessions.GET("/search", s.SearchSessions)
	sessions.GET("/:id", s.GetSession)
	sessions.POST("/:id/messages", s.SendMessage)

	api.POST("/research", s.StartResearch)

	kgr := api.Group("/knowledge")
	kgr.GET("/search", s.SearchKnowledge)
	kgr.GET("/graph", s.GetKnowledgeGraph)
	kgr.GET("/nodes/:id", s.GetNodeDetail)

	api.POST("/intent/classify", s.ClassifyIntent)
	api.GET("/autocomplete", s.GetAutocomplete)
	api.GET("/status", s.GetStatus)
	api.GET("/logs", s.GetLogs)

	return e
}
