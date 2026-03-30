# Zoro

Privacy-first research agent that builds a personal knowledge graph on your machine. Searches the web, extracts entities and relationships using local LLMs, and stores everything locally — your data never leaves your machine.

## Architecture

Zoro runs in two modes: **web** (development) and **desktop** (production).

```
Desktop mode (Wails native app)
  ├── Embedded Next.js frontend (static export in WebView)
  ├── Go backend (Echo v4, in-process via Wails AssetServer)
  ├── PostgreSQL with pgvector (Docker, port 5432)
  ├── Managed SearXNG subprocess (pip-installed venv, port 8888)
  └── Ollama (external — must be installed separately)

Web mode (development)
  ├── Next.js 16 frontend (port 3000, proxies /api/* to backend)
  ├── Go backend (Echo v4, port 8080)
  ├── PostgreSQL with pgvector via Docker (port 5432)
  ├── SearXNG via Docker (port 8888)
  └── Ollama (external)
```

**SDKs**:
- `github.com/urmzd/adk` — agent loop, provider interface, tool registry, Ollama adapter
- `github.com/urmzd/saige` — knowledge graph, PostgreSQL store (pgvector), extraction pipeline, embedder

## Prerequisites

### Desktop app

- [Go](https://go.dev) 1.25+
- [Node.js](https://nodejs.org) 22+
- [Ollama](https://ollama.ai)
- Python 3.10+ (for SearXNG auto-install)
- [Just](https://github.com/casey/just)

Docker is required for PostgreSQL. SearXNG is installed into a local Python venv automatically on first launch.

### Web development

- Everything above, plus [Docker](https://docs.docker.com/get-docker/) with Docker Compose

## Quick Start

### Desktop app (recommended)

```bash
just setup           # install frontend/Go deps, pull Ollama models
just build-desktop   # build the native desktop binary
./zoro-desktop       # launch (first run installs SearXNG; PostgreSQL must be running)
```

### Web development

```bash
just setup   # install deps, start Docker services, pull Ollama models
just dev     # start all services: Ollama, Docker, Go backend, Next.js frontend
```

## Commands

```
just setup           First-time setup (deps + models)
just dev             Web dev mode (Docker + Go + Next.js)
just build           Production web build
just stop            Stop Docker services
just build-desktop   Build native desktop app (Wails)
just dev-desktop     Run desktop app in dev mode
just check           Lint and typecheck (Go vet + Biome + tsc)
just generate        Regenerate frontend API client from OpenAPI spec
just pull <model>    Download an Ollama model
```

## Configuration

All environment variables are optional. Defaults work out of the box.

| Variable            | Default                  | Description                      |
|---------------------|--------------------------|----------------------------------|
| `OLLAMA_HOST`       | `http://localhost:11434` | Ollama server URL                |
| `OLLAMA_MODEL`      | `qwen3.5:4b`             | Main LLM for reasoning           |
| `OLLAMA_FAST_MODEL` | `qwen3.5:0.8b`           | Fast model for lightweight tasks |
| `EMBEDDING_MODEL`   | `nomic-embed-text`       | Model used for embeddings        |
| `POSTGRES_URL`      | `postgres://zoro:zoro@localhost:5432/zoro?sslmode=disable` | PostgreSQL connection URL |
| `SEARXNG_URL`       | (managed subprocess)     | Set to use external SearXNG      |
| `ZORO_DATA_DIR`     | `~/.config/zoro`         | App data directory               |

In web mode (`just dev`), `SEARXNG_URL` defaults to `http://127.0.0.1:8888` (Docker). PostgreSQL must always be running.

In desktop mode, SearXNG is managed as a subprocess unless overridden via env var.

## Desktop App Details

The desktop app bundles everything into a single workflow:

1. **PostgreSQL** with pgvector must be running via Docker (`docker compose up -d`). Data persists in a Docker volume.
2. **SearXNG** is installed via `pip` into a Python venv at `$ZORO_DATA_DIR/searxng-venv/` on first launch. Subsequent launches reuse the cached venv.
3. **Frontend** is embedded as a static export inside the Go binary via `go:embed`.
4. **API requests** from the WebView route through the Wails AssetServer to the Echo handler — no network ports needed for the frontend.

Logs are written to `$ZORO_DATA_DIR/zoro.log` and available via `GET /api/logs`.

## API

The OpenAPI spec at `api/openapi.yaml` is the single source of truth for backend types. Run `just generate` to regenerate the frontend API client after changes.

## Third-Party Licenses

See [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md) for attribution of bundled dependencies (including SearXNG, AGPL-3.0).

## License

[Apache License 2.0](LICENSE)
