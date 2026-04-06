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
	EmbeddingModel string
	PostgresURL     string
	SearXNGURL      string // empty = managed subprocess
	DataDir         string
}

func Load() *AppConfig {
	return &AppConfig{
		OllamaHost:      envOr("OLLAMA_HOST", "http://localhost:11434"),
		OllamaModel:     envOr("OLLAMA_MODEL", "gemma4:latest"),
		EmbeddingModel: envOr("EMBEDDING_MODEL", "nomic-embed-text"),
		PostgresURL:     envOr("POSTGRES_URL", "postgres://zoro:zoro@localhost:5432/zoro?sslmode=disable"),
		SearXNGURL:      os.Getenv("SEARXNG_URL"),
		DataDir:         envOr("ZORO_DATA_DIR", defaultDataDir()),
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
