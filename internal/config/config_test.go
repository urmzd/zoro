package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env vars to test defaults
	for _, key := range []string{
		"OLLAMA_HOST", "OLLAMA_MODEL",
		"EMBEDDING_MODEL", "POSTGRES_URL", "SEARXNG_URL", "ZORO_DATA_DIR",
	} {
		t.Setenv(key, "")
	}

	cfg := Load()

	if cfg.OllamaHost != "http://localhost:11434" {
		t.Errorf("OllamaHost = %q, want %q", cfg.OllamaHost, "http://localhost:11434")
	}
	if cfg.OllamaModel != "gemma4:latest" {
		t.Errorf("OllamaModel = %q, want %q", cfg.OllamaModel, "gemma4:latest")
	}
	if cfg.EmbeddingModel != "nomic-embed-text" {
		t.Errorf("EmbeddingModel = %q, want %q", cfg.EmbeddingModel, "nomic-embed-text")
	}
	if cfg.PostgresURL != "postgres://zoro:zoro@localhost:5432/zoro?sslmode=disable" {
		t.Errorf("PostgresURL = %q, want default", cfg.PostgresURL)
	}
	if cfg.SearXNGURL != "" {
		t.Errorf("SearXNGURL = %q, want empty", cfg.SearXNGURL)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "http://custom:9999")
	t.Setenv("OLLAMA_MODEL", "llama3:8b")
	t.Setenv("SEARXNG_URL", "http://searx:8080")
	t.Setenv("ZORO_DATA_DIR", "/tmp/zoro-test")

	cfg := Load()

	if cfg.OllamaHost != "http://custom:9999" {
		t.Errorf("OllamaHost = %q, want %q", cfg.OllamaHost, "http://custom:9999")
	}
	if cfg.OllamaModel != "llama3:8b" {
		t.Errorf("OllamaModel = %q, want %q", cfg.OllamaModel, "llama3:8b")
	}
	if cfg.SearXNGURL != "http://searx:8080" {
		t.Errorf("SearXNGURL = %q, want %q", cfg.SearXNGURL, "http://searx:8080")
	}
	if cfg.DataDir != "/tmp/zoro-test" {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, "/tmp/zoro-test")
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("TEST_ENV_OR_KEY", "custom")
	if got := envOr("TEST_ENV_OR_KEY", "default"); got != "custom" {
		t.Errorf("envOr with set var = %q, want %q", got, "custom")
	}

	t.Setenv("TEST_ENV_OR_KEY", "")
	if got := envOr("TEST_ENV_OR_KEY", "fallback"); got != "fallback" {
		t.Errorf("envOr with empty var = %q, want %q", got, "fallback")
	}
}

func TestDefaultDataDir(t *testing.T) {
	dir := defaultDataDir()
	if dir == "" {
		t.Fatal("defaultDataDir returned empty string")
	}
	if filepath.Base(dir) != "zoro" {
		t.Errorf("defaultDataDir base = %q, want %q", filepath.Base(dir), "zoro")
	}
}

func TestSearXNGSettingsEmbedded(t *testing.T) {
	if len(SearXNGSettings) == 0 {
		t.Fatal("SearXNGSettings is empty; embedded file missing")
	}
}

func TestLoad_DataDirFallback(t *testing.T) {
	// When ZORO_DATA_DIR is not set, should use UserConfigDir or TempDir
	t.Setenv("ZORO_DATA_DIR", "")
	cfg := Load()
	if cfg.DataDir == "" {
		t.Fatal("DataDir should not be empty when ZORO_DATA_DIR is unset")
	}

	configDir, err := os.UserConfigDir()
	if err == nil {
		expected := filepath.Join(configDir, "zoro")
		if cfg.DataDir != expected {
			t.Errorf("DataDir = %q, want %q", cfg.DataDir, expected)
		}
	}
}
