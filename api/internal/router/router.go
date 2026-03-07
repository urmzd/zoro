package router

import (
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/urmzd/zoro/api/internal/config"
	"github.com/urmzd/zoro/api/internal/handler"
	"github.com/urmzd/zoro/api/internal/service"
)

func New(cfg *config.Config, orchestrator *service.Orchestrator, knowledge service.KnowledgeStore, agent *service.Agent, ollama *service.OllamaClient, registry *service.ModelRegistry) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   parseOrigins(cfg.CORSOrigins),
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	healthHandler := handler.NewHealth()
	researchHandler := handler.NewResearch(orchestrator)
	knowledgeHandler := handler.NewKnowledge(knowledge)
	chatHandler := handler.NewChat(agent)
	intentHandler := handler.NewIntent(agent)
	autocompleteHandler := handler.NewAutocomplete(ollama, registry)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Get("/health", healthHandler.Check)

	r.Route("/api", func(r chi.Router) {
		r.Post("/research", researchHandler.Start)
		r.Get("/research/{id}", researchHandler.Get)
		r.Get("/research/{id}/stream", researchHandler.Stream)

		r.Get("/knowledge/search", knowledgeHandler.Search)
		r.Get("/knowledge/graph", knowledgeHandler.Graph)
		r.Get("/knowledge/node/{id}", knowledgeHandler.Node)

		r.Post("/intent", intentHandler.Classify)
		r.Get("/autocomplete", autocompleteHandler.Suggest)

		r.Get("/chat/sessions", chatHandler.ListSessions)
		r.Post("/chat/sessions", chatHandler.CreateSession)
		r.Get("/chat/sessions/{id}", chatHandler.GetSession)
		r.Post("/chat/sessions/{id}/messages", chatHandler.SendMessage)
	})

	return r
}

func parseOrigins(s string) []string {
	origins := strings.Split(s, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
}
