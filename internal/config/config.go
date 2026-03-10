package config

import "os"

type AppConfig struct {
	OllamaHost     string
	OllamaModel    string
	OllamaFastModel string
	EmbeddingModel string
	SurrealDBURL   string
}

func Load() *AppConfig {
	return &AppConfig{
		OllamaHost:      envOr("OLLAMA_HOST", "http://localhost:11434"),
		OllamaModel:     envOr("OLLAMA_MODEL", "qwen3.5:4b"),
		OllamaFastModel: envOr("OLLAMA_FAST_MODEL", "qwen3.5:0.8b"),
		EmbeddingModel:  envOr("EMBEDDING_MODEL", "nomic-embed-text"),
		SurrealDBURL:    envOr("SURREALDB_URL", "ws://localhost:8000"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
