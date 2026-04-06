# AGENTS.md

## Project Overview

Zoro is a privacy-first research agent that builds a personal knowledge graph locally. It searches the web via SearXNG, extracts entities and relationships using local LLMs (Ollama), and stores everything in PostgreSQL with pgvector. All data stays on the user's machine.

**Stack**: Go 1.25, PostgreSQL (pgvector), SearXNG, Ollama.

## Architecture

Zoro exposes capabilities through a CLI.

```
Terminal
  └── zoro <command>
        │
        ├── Chat agent (multi-turn, session persistence)
        ├── Research pipeline (web search → ingest → LLM synthesis)
        ├── Knowledge graph (PostgreSQL + pgvector)
        ├── SearXNG (web search, Docker or managed subprocess)
        └── Ollama (local LLMs)
```

## Setup Commands

```bash
just setup           # Install deps, start Docker, pull Ollama models
just dev             # Start Docker + Ollama services
just stop            # Stop Docker services
```

**Prerequisites**: Go 1.25+, Docker with Docker Compose, Ollama, Just

**Required Ollama models** (pulled by `just setup`):
- `qwen3.5:4b` — main LLM
- `qwen3.5:0.8b` — fast model (intent classification, autocomplete)
- `nomic-embed-text` — embeddings

## Development Workflow

- `just dev` starts Docker services and Ollama.
- CLI subcommands (`chat`, `research`, `search`, `knowledge`, `graph`, `version`) share the same internals.
- SearXNG runs as a Docker container (default) or managed subprocess when `SEARXNG_URL` is unset.

## Environment Variables

All optional; defaults work for local development.

| Variable | Default | Purpose |
|----------|---------|---------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODEL` | `qwen3.5:4b` | Main LLM model |
| `OLLAMA_FAST_MODEL` | `qwen3.5:0.8b` | Fast model (autocomplete/intent) |
| `EMBEDDING_MODEL` | `nomic-embed-text` | Embedding model |
| `POSTGRES_URL` | `postgres://zoro:zoro@localhost:5432/zoro?sslmode=disable` | PostgreSQL connection URL |
| `SEARXNG_URL` | (empty = managed subprocess) | Set to use external SearXNG |
| `ZORO_DATA_DIR` | `~/.config/zoro` | App data dir (binaries, DB, venv, logs) |

When `SEARXNG_URL` is empty in `just dev`, it defaults to `http://127.0.0.1:8888` (Docker). PostgreSQL must always be running (Docker).

## Subprocess Management

`internal/subprocess/` manages SearXNG as a child process:

### SearXNG (`searxng.go`)
- Creates a Python venv at `$ZORO_DATA_DIR/searxng-venv/`
- Installs SearXNG from GitHub master via pip (with setuptools, wheel, msgspec, typing_extensions)
- Runs Flask server on port 8888
- Settings file embedded in Go binary (`internal/config/searxng-settings.yml`), written to `$ZORO_DATA_DIR/searxng/settings.yml`

## Code Quality

```bash
just check            # go vet ./...
just test             # go test -race -count=1 ./...
just lint             # golangci-lint run ./...
just vuln             # govulncheck ./...
just ci               # Full CI gate (check + test + build)
```

- **Go**: `go vet`, golangci-lint, govulncheck
- **Tests**: Unit tests in `*_test.go` files across packages (config, graph, searcher, tools)
- **CI**: GitHub Actions runs vet, test, lint, vuln check, and build on PRs

## Build

```bash
just build            # go build -o zoro .
just install          # go install .
```

CI pipeline (`.github/workflows/ci.yml`) runs check, test, lint, vuln, then build.
Releases use semantic versioning via `.github/workflows/release.yml`.

## Code Conventions

### Go

- Entry point: `main.go` — CLI dispatch via os.Args subcommand routing
- CLI commands: `cmd/chat.go`, `cmd/research.go`, `cmd/search.go`, `cmd/knowledge.go`, `cmd/graph.go`, `cmd/version.go`
- Shared wiring: `internal/app/wire.go` — `WireComponents(ctx, cfg, opts)` returns configured components + cleanup
- All domain code in `internal/` (agent, app, config, events, graph, models, orchestrator, searcher, subprocess, tools)
- Tools implement the `saige/agent/types.Tool` interface: `Definition()` + `Execute()`
- PostgreSQL queries use `pgx` via `pgxpool`
- No ORM; raw SQL in `internal/events/store.go`
- Searcher accepts a base URL: `searcher.New(baseURL string)`

### Naming

- **Go**: PascalCase types, camelCase locals, snake_case in SQL
- **Commits**: conventional commits (feat, fix, chore, etc.) — enforced by semantic release

## CLI Commands

| Command | Description |
|---------|-------------|
| `chat` | Multi-turn conversation with the agent. Pass `-s SESSION_ID` to continue a session. |
| `research` | Deep research pipeline: web search, knowledge graph ingestion, and LLM synthesis. |
| `search` | Search the web via SearXNG. Use `-json` for machine-readable output. |
| `knowledge search` | Query the knowledge graph for stored facts and entities. |
| `knowledge store` | Ingest text into the knowledge graph, extracting entities and relationships. |
| `graph` | Visualize the knowledge graph structure (text, DOT, or JSON format). |
| `version` | Print version. |

## Key Dependencies

- **`github.com/urmzd/saige`**: Agent loop, knowledge graph, pgvector store, extraction pipeline, Ollama adapter

## Troubleshooting

- If Ollama requests time out (30s), the model may be cold-loading. First request after model switch is slow.
- Docker uses Colima on macOS. If Docker commands fail, run `colima start`.
- Port conflicts on 8080: check for leftover `zoro` processes with `lsof -i :8080`.
- SearXNG warning about `limiter.toml` is non-critical and can be ignored.
- SearXNG first install takes ~30s (pip install from GitHub). Subsequent launches reuse the cached venv (~1s).
- If SearXNG pip install fails, delete `$ZORO_DATA_DIR/searxng-venv/` and relaunch.
