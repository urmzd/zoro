package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	// Web mode defaults: expect external SurrealDB and SearXNG (Docker)
	if cfg.SurrealDBURL == "" {
		cfg.SurrealDBURL = "ws://localhost:8000"
	}
	if cfg.SearXNGURL == "" {
		cfg.SearXNGURL = "http://127.0.0.1:8888"
	}

	e, cleanup, err := app.Wire(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to wire app: %v", err)
	}
	defer cleanup()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		cancel()
		e.Close()
	}()

	addr := ":" + cfg.Port
	log.Printf("Zoro backend starting on %s", addr)
	if err := e.Start(addr); err != nil {
		log.Printf("server stopped: %v", err)
	}
}
