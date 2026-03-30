# AGENTS.md

## Project Overview

Zoro is a privacy-first research agent that builds a personal knowledge graph locally. It searches the web via SearXNG, extracts entities and relationships using local LLMs (Ollama), and stores everything in PostgreSQL with pgvector. All data stays on the user's machine.

**Stack**: Go 1.25 + Echo v4 backend, Next.js 16 + React 19 frontend, Wails v2 desktop shell, PostgreSQL (pgvector), SearXNG, Ollama.

## Architecture

Zoro has two run modes:

```
Desktop mode (Wails — requires Docker for PostgreSQL)
  ├── Wails WebView renders embedded static Next.js export
  ├── Echo v4 serves API via Wails AssetServer (no network port)
  ├── internal/subprocess/ manages SearXNG lifecycle
  │   └── SearXNG: pip-installed Python venv (port 8888)
  ├── PostgreSQL with pgvector (Docker, port 5432)
  └── Ollama (external, must be running)

Web mode (development — uses Docker)
  ├── Next.js dev server (port 3000) proxies /api/* → Go backend
  ├── Echo v4 backend (port 8080)
  ├── PostgreSQL with pgvector via Docker (port 5432)
  ├── SearXNG via Docker (port 8888)
  └── Ollama (external)
```

## Setup Commands

```bash
# Desktop app
just build-desktop   # Build native binary (embeds frontend)
./zoro-desktop       # First run installs SearXNG; PostgreSQL must be running

# Web development
just setup           # Install deps, start Docker, pull Ollama models
just dev             # Start Go backend + Next.js dev server
just stop            # Stop Docker services
```

**Desktop prerequisites**: Go 1.25+, Node.js 22+, Ollama, Python 3.10+, Just
**Web prerequisites**: All the above + Docker with Docker Compose

**Required Ollama models** (pulled by `just setup`):
- `qwen3.5:4b` — main LLM
- `qwen3.5:0.8b` — fast model (intent classification, autocomplete)
- `nomic-embed-text` — embeddings

## Development Workflow

- `just dev` starts web mode. Go backend on `:8080`, Next.js on `:3000`.
- `just dev-desktop` builds frontend + runs the Wails desktop app with dev tags.
- Frontend proxies `/api/*` to the Go backend via `next.config.ts` rewrites (web mode only).
- In desktop mode, `NEXT_BUILD_TARGET=desktop` triggers a static export (`output: 'export'`).
- SSE streaming: backend sends `data: {"type":"...","data":{...}}\n\n`; frontend consumes via `fetch` + `ReadableStream`.
- API spec lives in `api/openapi.yaml` (OpenAPI 3.1, single source of truth for types).
- Run `just generate` to regenerate the frontend API client (uses `oag` with `.urmzd.oag.yaml`).

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

In web mode (`main.go`), empty `SEARXNG_URL` defaults to `http://127.0.0.1:8888` (Docker). In desktop mode (`cmd/desktop/main.go`), empty `SEARXNG_URL` triggers subprocess management. PostgreSQL must always be running (Docker).

## Subprocess Management

`internal/subprocess/` manages SearXNG as a child process for the desktop app:

### SearXNG (`searxng.go`)
- Creates a Python venv at `$ZORO_DATA_DIR/searxng-venv/`
- Installs SearXNG from GitHub master via pip (with setuptools, wheel, msgspec, typing_extensions)
- Runs Flask server on port 8888
- Settings file embedded in Go binary (`internal/config/searxng-settings.yml`), written to `$ZORO_DATA_DIR/searxng/settings.yml`

## Wails Desktop Integration

- Entry point: `cmd/desktop/main.go`
- Build tags: `-tags desktop,production` (release) or `-tags desktop,dev` (development)
- Frontend assets embedded via `//go:embed all:assets` (copied from `frontend/out/`)
- macOS needs `CGO_LDFLAGS="-framework UniformTypeIdentifiers"` for linking
- Wails `AssetServer.Handler` routes to the Echo instance — API calls work without a network port
- CORS origins include `wails://wails`, `http://wails.localhost`, `http://localhost` for WebView
- Title bar uses `TitleBarHiddenInset` with a 32px drag region in the frontend layout
- Logs written to `$ZORO_DATA_DIR/zoro.log`

## Code Quality

```bash
just check            # Run all checks (Go + frontend)
just check-go         # go vet ./...
just check-frontend   # biome check . && tsc --noEmit
```

- **Go**: `go vet` (no additional linters configured)
- **Frontend**: Biome for linting + formatting, TypeScript strict mode for type checking
- **No test suite**: No test files or test frameworks exist yet

## Build

```bash
just build            # Web: go build -o zoro . && next build
just build-desktop    # Desktop: static export + go build with Wails tags
```

CI pipeline (`.github/workflows/ci.yml`) runs `check-go`, `check-frontend`, then `build`.
Releases use semantic versioning via `.github/workflows/release.yml`.

## Code Conventions

### Go Backend

- Entry point: `main.go` (web mode), `cmd/desktop/main.go` (desktop mode)
- Shared wiring: `internal/app/wire.go` — `Wire(ctx, cfg)` returns configured Echo + cleanup func
- All domain code in `internal/` (agent, app, config, events, models, orchestrator, searcher, server, subprocess, tools)
- Tools implement the `adk.Tool` interface: `Definition()` + `Execute()`
- HTTP handlers in `internal/server/handlers_*.go`, grouped by domain
- SSE streaming via `internal/server/sse.go`
- PostgreSQL queries use `pgx` via `pgxpool`
- No ORM; raw SQL in `internal/events/store.go`
- Searcher accepts a base URL: `searcher.New(baseURL string)`

### Frontend

- Pages: `frontend/src/app/{page,chat,research,knowledge}/page.tsx`
- Components: `frontend/src/components/{chat,graph,knowledge,nav,research,timeline,ui}/`
- State: Zustand stores in `frontend/src/lib/stores/`
- API client: `frontend/src/app/lib/api.ts`
- SSE hooks: `use-chat-stream.ts`, `use-research-stream.ts`, `use-knowledge-chat-stream.ts`
- UI primitives: shadcn/ui in `frontend/src/components/ui/`
- Biome config: 2-space indent, double quotes, semicolons, line width 100
- Path alias: `@/*` maps to `./src/*`
- Pages using `useSearchParams()` must be wrapped in `<Suspense>`
- `react-force-graph-2d` loaded via `LazyForceGraph2D` wrapper (deferred client-side import)
- `next.config.ts` conditionally exports static build when `NEXT_BUILD_TARGET=desktop`

### Naming

- **Go**: PascalCase types, camelCase locals, snake_case in SQL
- **Frontend**: PascalCase components, camelCase functions/variables
- **API routes**: kebab-case (`/api/sessions`, `/api/knowledge/search`)
- **Commits**: conventional commits (feat, fix, chore, etc.) — enforced by semantic release

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/sessions` | Create chat session |
| GET | `/api/sessions` | List sessions |
| GET | `/api/sessions/search` | Search sessions by text |
| GET | `/api/sessions/{id}` | Get session with messages |
| POST | `/api/sessions/{id}/messages` | Send message (SSE stream) |
| POST | `/api/research` | Start research (SSE stream) |
| GET | `/api/knowledge/search` | Search knowledge graph |
| GET | `/api/knowledge/graph` | Get full graph data |
| GET | `/api/knowledge/nodes/{id}` | Get node with neighbors |
| POST | `/api/intent/classify` | Classify query intent |
| GET | `/api/autocomplete` | Autocomplete suggestions |
| GET | `/api/status` | Service readiness (postgres, searxng, ollama) |
| GET | `/api/logs?lines=100` | App log tail |

## Key Dependencies

- **`github.com/urmzd/adk`**: Agent development kit — agent loop, provider interface, tool registry, Ollama adapter, SSE streaming
- **`github.com/urmzd/saige`**: Knowledge graph, PostgreSQL store (pgvector), entity extraction, embeddings
- **`github.com/wailsapp/wails/v2`**: Native desktop shell — WebView, asset server, macOS/Windows/Linux
- SDKs may use local `replace` directives in `go.mod` during development

## Troubleshooting

- If Ollama requests time out (30s), the model may be cold-loading. First request after model switch is slow.
- Docker uses Colima on macOS. If Docker commands fail, run `colima start`. Docker is only needed for web dev mode.
- Port conflicts on 8080: check for leftover `zoro` processes with `lsof -i :8080`.
- SearXNG warning about `limiter.toml` is non-critical and can be ignored.
- SearXNG first install takes ~30s (pip install from GitHub). Subsequent launches reuse the cached venv (~1s).
- Desktop app logs: `cat "$(python3 -c 'import os; print(os.path.join(os.path.expanduser("~"), "Library", "Application Support", "zoro", "zoro.log"))')"`
- macOS linker errors with Wails: ensure `CGO_LDFLAGS="-framework UniformTypeIdentifiers"` is set (handled by justfile).
- If SearXNG pip install fails, delete `$ZORO_DATA_DIR/searxng-venv/` and relaunch.
