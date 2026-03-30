package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urmzd/saige/agent/provider/ollama"
	"github.com/urmzd/saige/knowledge"
	kgtypes "github.com/urmzd/saige/knowledge/types"
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

// Components holds all wired dependencies.
type Components struct {
	MCP          *gomcp.Server
	Agent        *agent.Agent
	Orchestrator *orchestrator.Orchestrator
	Searcher     *searcher.Searcher
	Graph        kgtypes.Graph
	Cleanup      func()
}

// WireOpts controls which subsystems to initialize.
type WireOpts struct {
	NeedAgent        bool
	NeedOrchestrator bool
	NeedSearcher     bool
	NeedMCP          bool
	NeedGraph        bool
}

// WireAll returns opts that initialize everything.
func WireAll() WireOpts {
	return WireOpts{
		NeedAgent:        true,
		NeedOrchestrator: true,
		NeedSearcher:     true,
		NeedMCP:          true,
	}
}

// WireComponents creates dependencies based on the requested opts.
func WireComponents(ctx context.Context, cfg *config.AppConfig, opts WireOpts) (*Components, error) {
	var cleanups []func()
	c := &Components{}

	needDB := opts.NeedAgent || opts.NeedOrchestrator || opts.NeedMCP || opts.NeedGraph
	needOllama := opts.NeedAgent || opts.NeedOrchestrator || opts.NeedMCP
	needSearcher := opts.NeedSearcher || opts.NeedOrchestrator || opts.NeedAgent || opts.NeedMCP

	// SearXNG
	var s *searcher.Searcher
	if needSearcher {
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
		s = searcher.New(searxngURL)
		c.Searcher = s
	}

	// PostgreSQL
	var pool *pgxpool.Pool
	if needDB {
		if err := ensureExtensions(ctx, cfg.PostgresURL); err != nil {
			runCleanups(cleanups)
			return nil, fmt.Errorf("ensure extensions: %w", err)
		}

		var err error
		pool, err = postgres.NewPool(ctx, postgres.Config{URL: cfg.PostgresURL})
		if err != nil {
			runCleanups(cleanups)
			return nil, fmt.Errorf("connect postgres: %w", err)
		}
		cleanups = append(cleanups, func() { pool.Close() })

		if err := postgres.RunMigrations(ctx, pool, postgres.MigrationOptions{}); err != nil {
			runCleanups(cleanups)
			return nil, fmt.Errorf("run migrations: %w", err)
		}
	}

	// Ollama
	var ollamaClient *ollama.Client
	var adapter *ollama.Adapter
	if needOllama {
		ollamaClient = ollama.NewClient(cfg.OllamaHost, cfg.OllamaModel, cfg.EmbeddingModel)
		adapter = ollama.NewAdapter(ollamaClient)
	}

	// Knowledge graph
	var graph kgtypes.Graph
	if needDB && needOllama {
		embedder := knowledge.NewOllamaEmbedder(ollamaClient)
		extractor := knowledge.NewOllamaExtractor(ollamaClient)

		g, err := knowledge.NewGraph(ctx,
			knowledge.WithPostgres(pool),
			knowledge.WithExtractor(extractor),
			knowledge.WithEmbedder(embedder),
		)
		if err != nil {
			runCleanups(cleanups)
			return nil, err
		}
		graph = g
		cleanups = append(cleanups, func() { graph.Close(ctx) })
	} else if needDB && opts.NeedGraph {
		// Read-only graph: no embedder/extractor needed.
		g, err := knowledge.NewGraph(ctx, knowledge.WithPostgres(pool))
		if err != nil {
			runCleanups(cleanups)
			return nil, err
		}
		graph = g
		cleanups = append(cleanups, func() { graph.Close(ctx) })
	}
	c.Graph = graph

	// Event store
	var es *events.Store
	if needDB {
		es = events.New(pool)
		if err := es.EnsureSchema(ctx); err != nil {
			log.Printf("event schema warning: %v", err)
		}
	}

	// Agent
	if opts.NeedAgent || opts.NeedMCP {
		webSearch := tools.NewWebSearchTool(s, graph)
		searchKG := tools.NewSearchKnowledgeTool(graph)
		storeKG := tools.NewStoreKnowledgeTool(graph)
		getGraph := tools.NewGetGraphTool(graph)
		c.Agent = agent.New(adapter, webSearch, searchKG, storeKG, getGraph, es)
	}

	// Orchestrator
	if opts.NeedOrchestrator || opts.NeedMCP {
		c.Orchestrator = orchestrator.New(graph, adapter, s)
	}

	// MCP server
	if opts.NeedMCP {
		c.MCP = mcp.NewServer(c.Agent, c.Orchestrator, graph, s)
	}

	c.Cleanup = func() { runCleanups(cleanups) }
	return c, nil
}

// Wire creates all dependencies and returns a configured MCP server.
func Wire(ctx context.Context, cfg *config.AppConfig) (*gomcp.Server, func(), error) {
	c, err := WireComponents(ctx, cfg, WireAll())
	if err != nil {
		return nil, nil, err
	}
	return c.MCP, c.Cleanup, nil
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
