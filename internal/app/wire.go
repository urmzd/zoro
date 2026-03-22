package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/urmzd/adk/provider/ollama"
	kg "github.com/urmzd/kgdk"
	kgsurrealdb "github.com/urmzd/kgdk/surrealdb"
	"github.com/urmzd/zoro/internal/agent"
	"github.com/urmzd/zoro/internal/config"
	"github.com/urmzd/zoro/internal/events"
	"github.com/urmzd/zoro/internal/orchestrator"
	"github.com/urmzd/zoro/internal/searcher"
	"github.com/urmzd/zoro/internal/server"
	"github.com/urmzd/zoro/internal/subprocess"
	"github.com/urmzd/zoro/internal/tools"
)

// Wire creates all dependencies and returns a configured Echo instance.
// The returned cleanup function closes connections and stops subprocesses.
func Wire(ctx context.Context, cfg *config.AppConfig) (*echo.Echo, func(), error) {
	var cleanups []func()

	// SurrealDB: managed subprocess or external
	surrealURL := cfg.SurrealDBURL
	if surrealURL == "" {
		proc, err := subprocess.StartSurreal(ctx, cfg.DataDir, 8765)
		if err != nil {
			return nil, nil, fmt.Errorf("start surrealdb: %w", err)
		}
		surrealURL = proc.URL()
		cleanups = append(cleanups, func() { proc.Stop() })
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

	ollamaClient := ollama.NewClient(cfg.OllamaHost, cfg.OllamaModel, cfg.EmbeddingModel)
	adapter := ollama.NewAdapter(ollamaClient)

	embedder := kg.NewOllamaEmbedder(ollamaClient)
	extractor := kg.NewOllamaExtractor(ollamaClient)

	store, err := kgsurrealdb.NewStore(ctx, kgsurrealdb.StoreConfig{
		URL:       surrealURL,
		Namespace: "zoro",
		Database:  "zoro",
		Username:  "root",
		Password:  "root",
	})
	if err != nil {
		runCleanups(cleanups)
		return nil, nil, fmt.Errorf("connect surrealdb: %w", err)
	}
	cleanups = append(cleanups, func() { store.Close(ctx) })

	graph, err := kg.NewGraph(ctx,
		kg.WithStore(store),
		kg.WithExtractor(extractor),
		kg.WithEmbedder(embedder),
	)
	if err != nil {
		runCleanups(cleanups)
		return nil, nil, err
	}

	es := events.New(ctx, store.DB())
	if err := es.EnsureSchema(); err != nil {
		log.Printf("event schema warning: %v", err)
	}

	s := searcher.New(searxngURL)

	webSearch := tools.NewWebSearchTool(s)
	searchKG := tools.NewSearchKnowledgeTool(graph)
	storeKG := tools.NewStoreKnowledgeTool(graph)

	ag := agent.New(adapter, webSearch, searchKG, storeKG, cfg.OllamaFastModel, es)
	orch := orchestrator.New(graph, adapter, s)

	srv := server.New(ag, orch, graph, adapter)
	srv.SetServiceStatus(server.ServiceStatus{
		SurrealDB: true, // we got here, so SurrealDB is connected
		SearXNG:   searxngURL != "",
		Ollama:    checkOllama(cfg.OllamaHost),
	})
	e := srv.Setup()

	cleanup := func() {
		runCleanups(cleanups)
	}

	return e, cleanup, nil
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

func checkOllama(host string) bool {
	resp, err := http.Get(host + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func runCleanups(fns []func()) {
	for i := len(fns) - 1; i >= 0; i-- {
		fns[i]()
	}
}
