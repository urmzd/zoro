package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/urmzd/saige/agent/provider/ollama"
	"github.com/urmzd/saige/knowledge"
	"github.com/urmzd/saige/postgres"
	"github.com/urmzd/zoro/internal/agent"
	"github.com/urmzd/zoro/internal/config"
	"github.com/urmzd/zoro/internal/events"
	"github.com/urmzd/zoro/internal/mcp"
	"github.com/urmzd/zoro/internal/orchestrator"
	"github.com/urmzd/zoro/internal/searcher"
	"github.com/urmzd/zoro/internal/subprocess"
	"github.com/urmzd/zoro/internal/tools"
)

// Wire creates all dependencies and returns a configured MCP server.
func Wire(ctx context.Context, cfg *config.AppConfig) (*mcpserver.MCPServer, func(), error) {
	var cleanups []func()

	// PostgreSQL
	if err := ensureExtensions(ctx, cfg.PostgresURL); err != nil {
		return nil, nil, fmt.Errorf("ensure extensions: %w", err)
	}

	pool, err := postgres.NewPool(ctx, postgres.Config{URL: cfg.PostgresURL})
	if err != nil {
		return nil, nil, fmt.Errorf("connect postgres: %w", err)
	}
	cleanups = append(cleanups, func() { pool.Close() })

	if err := postgres.RunMigrations(ctx, pool, postgres.MigrationOptions{}); err != nil {
		runCleanups(cleanups)
		return nil, nil, fmt.Errorf("run migrations: %w", err)
	}

	// SearXNG: managed subprocess or external
	searxngURL := cfg.SearXNGURL
	if searxngURL == "" {
		settingsPath, err := writeSearXNGSettings(cfg.DataDir)
		if err != nil {
			log.Printf("warning: failed to write searxng settings: %v", err)
		} else {
			proc, err := subprocess.StartSearXNG(ctx, cfg.DataDir, 8888, settingsPath)
			if err != nil {
				log.Printf("warning: failed to start searxng: %v (web search will be unavailable)", err)
				searxngURL = "http://127.0.0.1:8888"
			} else {
				searxngURL = proc.URL()
				cleanups = append(cleanups, func() { proc.Stop() })
			}
		}
		if searxngURL == "" {
			searxngURL = "http://127.0.0.1:8888"
		}
	}

	// Single Ollama client for both agent and knowledge graph
	ollamaClient := ollama.NewClient(cfg.OllamaHost, cfg.OllamaModel, cfg.EmbeddingModel)
	adapter := ollama.NewAdapter(ollamaClient)

	embedder := knowledge.NewOllamaEmbedder(ollamaClient)
	extractor := knowledge.NewOllamaExtractor(ollamaClient)

	// Knowledge graph
	graph, err := knowledge.NewGraph(ctx,
		knowledge.WithPostgres(pool),
		knowledge.WithExtractor(extractor),
		knowledge.WithEmbedder(embedder),
	)
	if err != nil {
		runCleanups(cleanups)
		return nil, nil, err
	}
	cleanups = append(cleanups, func() { graph.Close(ctx) })

	// Chat session store
	es := events.New(pool)
	if err := es.EnsureSchema(ctx); err != nil {
		log.Printf("event schema warning: %v", err)
	}

	s := searcher.New(searxngURL)

	webSearch := tools.NewWebSearchTool(s, graph)
	searchKG := tools.NewSearchKnowledgeTool(graph)
	storeKG := tools.NewStoreKnowledgeTool(graph)

	ag := agent.New(adapter, webSearch, searchKG, storeKG, es)
	orch := orchestrator.New(graph, adapter, s)

	srv := mcp.NewServer(ag, orch, graph, s)

	cleanup := func() {
		runCleanups(cleanups)
	}

	return srv, cleanup, nil
}

func writeSearXNGSettings(dataDir string) (string, error) {
	settingsDir := filepath.Join(dataDir, "searxng")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(settingsDir, "settings.yml")
	if err := os.WriteFile(p, config.SearXNGSettings, 0o644); err != nil {
		return "", err
	}
	return p, nil
}

func ensureExtensions(ctx context.Context, dsn string) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	for _, ext := range []string{"vector", "pg_trgm"} {
		if _, err := conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS "+ext); err != nil {
			return fmt.Errorf("create extension %s: %w", ext, err)
		}
	}
	return nil
}

func runCleanups(fns []func()) {
	for i := len(fns) - 1; i >= 0; i-- {
		fns[i]()
	}
}
