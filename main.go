package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urmzd/adk/provider/ollama"
	kg "github.com/urmzd/kgdk"
	kgsurrealdb "github.com/urmzd/kgdk/surrealdb"
	"github.com/urmzd/zoro/internal/agent"
	"github.com/urmzd/zoro/internal/config"
	"github.com/urmzd/zoro/internal/events"
	"github.com/urmzd/zoro/internal/orchestrator"
	"github.com/urmzd/zoro/internal/searcher"
	"github.com/urmzd/zoro/internal/server"
	"github.com/urmzd/zoro/internal/tools"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	// Ollama client + adapter
	ollamaClient := ollama.NewClient(cfg.OllamaHost, cfg.OllamaModel, cfg.EmbeddingModel)
	adapter := ollama.NewAdapter(ollamaClient)

	// Embedder + extractor for knowledge graph
	embedder := kg.NewOllamaEmbedder(ollamaClient)
	extractor := kg.NewOllamaExtractor(ollamaClient)

	// SurrealDB store (created directly for DB connection sharing)
	store, err := kgsurrealdb.NewStore(ctx, kgsurrealdb.StoreConfig{
		URL:       cfg.SurrealDBURL,
		Namespace: "zoro",
		Database:  "zoro",
		Username:  "root",
		Password:  "root",
	})
	if err != nil {
		log.Fatalf("failed to connect to knowledge graph: %v", err)
	}
	defer store.Close(ctx)

	// Knowledge graph (wraps store with extraction/dedup engine)
	graph, err := kg.NewGraph(ctx,
		kg.WithStore(store),
		kg.WithExtractor(extractor),
		kg.WithEmbedder(embedder),
	)
	if err != nil {
		log.Fatalf("failed to create knowledge graph: %v", err)
	}

	// Event store shares the same SurrealDB connection
	es := events.New(ctx, store.DB())
	if err := es.EnsureSchema(); err != nil {
		log.Printf("event schema warning: %v", err)
	}

	// Searcher
	s := searcher.New()

	// Tool wrappers
	webSearch := tools.NewWebSearchTool(s)
	searchKG := tools.NewSearchKnowledgeTool(graph)
	storeKG := tools.NewStoreKnowledgeTool(graph)

	// Agent
	ag := agent.New(adapter, webSearch, searchKG, storeKG, cfg.OllamaFastModel, es)

	// Orchestrator
	orch := orchestrator.New(graph, adapter, s)

	// HTTP server
	srv := server.New(ag, orch, graph, adapter)
	e := srv.Setup()

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		cancel()
		e.Close()
	}()

	log.Println("Zoro backend starting on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Printf("server stopped: %v", err)
	}
}
