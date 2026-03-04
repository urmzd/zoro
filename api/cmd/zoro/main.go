package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	_ "github.com/urmzd/zoro/api/docs"

	"github.com/urmzd/zoro/api/internal/config"
	"github.com/urmzd/zoro/api/internal/router"
	"github.com/urmzd/zoro/api/internal/service"
)

//	@title			Zoro API
//	@version		1.0
//	@description	Research orchestration API with knowledge graph integration.
//	@host			localhost:8080
//	@BasePath		/
func main() {
	cfg := config.Load()

	ctx := context.Background()

	// Create Neo4j driver
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4jURI, neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	if err != nil {
		log.Fatalf("neo4j driver: %v", err)
	}
	defer driver.Close(ctx)

	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("neo4j connectivity: %v", err)
	}
	log.Println("Connected to Neo4j")

	ollamaClient := service.NewOllamaClient(cfg.OllamaHost, cfg.OllamaModel, cfg.EmbeddingModel)
	knowledge := service.NewNeo4jKnowledgeStore(driver, ollamaClient)

	if err := knowledge.EnsureSchema(ctx); err != nil {
		log.Printf("schema setup warning: %v", err)
	}
	log.Println("Neo4j schema ensured")

	searcher := service.NewSearcher(cfg.SearXNGURL)
	orchestrator := service.NewOrchestrator(knowledge, ollamaClient, searcher)

	registry := service.NewModelRegistry(cfg.OllamaModel, cfg.OllamaFastModel, cfg.EmbeddingModel)
	toolRegistry := service.NewToolRegistry(searcher, knowledge)
	agent := service.NewAgent(ollamaClient, toolRegistry, registry)

	r := router.New(cfg, orchestrator, knowledge, agent, ollamaClient, registry)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // no timeout for SSE
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Zoro API listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
