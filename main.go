package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/urmzd/zoro/internal/app"
	"github.com/urmzd/zoro/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	if cfg.SearXNGURL == "" {
		cfg.SearXNGURL = "http://127.0.0.1:8888"
	}

	srv, cleanup, err := app.Wire(ctx, cfg)
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
		os.Exit(0)
	}()

	log.Println("Zoro MCP server starting on stdio")
	if err := mcpserver.ServeStdio(srv); err != nil {
		log.Fatalf("mcp server error: %v", err)
	}
}
