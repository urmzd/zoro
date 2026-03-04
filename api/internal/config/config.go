package config

import "os"

type Config struct {
	Neo4jURI       string
	Neo4jUser      string
	Neo4jPassword  string
	OllamaHost     string
	OllamaModel    string
	EmbeddingModel string
	CORSOrigins    string
}

func Load() *Config {
	return &Config{
		Neo4jURI:       getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:      getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:  getEnv("NEO4J_PASSWORD", "zoro_dev_password"),
		OllamaHost:     getEnv("OLLAMA_HOST", "http://localhost:11434"),
		OllamaModel:    getEnv("OLLAMA_MODEL", "qwen3.5:4b"),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", "nomic-embed-text"),
		CORSOrigins:    getEnv("CORS_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
