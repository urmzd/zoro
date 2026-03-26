package config

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed searxng-settings.yml
var SearXNGSettings []byte

type AppConfig struct {
	OllamaHost      string
	OllamaModel     string
	OllamaFastModel string
	EmbeddingModel  string
	SurrealDBURL    string // empty = managed subprocess
	SurrealDBUser   string
	SurrealDBPass   string
	SearXNGURL      string // empty = managed subprocess
	DataDir         string
	Port            string
}

func Load() *AppConfig {
	return &AppConfig{
		OllamaHost:      envOr("OLLAMA_HOST", "http://localhost:11434"),
		OllamaModel:     envOr("OLLAMA_MODEL", "qwen3.5:4b"),
		OllamaFastModel: envOr("OLLAMA_FAST_MODEL", "qwen3.5:0.8b"),
		EmbeddingModel:  envOr("EMBEDDING_MODEL", "nomic-embed-text"),
		SurrealDBURL:    os.Getenv("SURREALDB_URL"),
		SurrealDBUser:   envOr("SURREALDB_USER", "root"),
		SurrealDBPass:   envOr("SURREALDB_PASS", "root"),
		SearXNGURL:      os.Getenv("SEARXNG_URL"),
		DataDir:         envOr("ZORO_DATA_DIR", defaultDataDir()),
		Port:            envOr("PORT", "8080"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "zoro")
}
