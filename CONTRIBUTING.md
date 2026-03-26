# Contributing to Zoro

Thanks for your interest in contributing to Zoro. This guide covers how to set up a development environment and submit changes.

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 22+](https://nodejs.org/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Ollama](https://ollama.ai)
- [Just](https://github.com/casey/just)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/urmzd/zoro.git
cd zoro

# First-time setup: installs deps, starts Docker (SurrealDB + SearXNG), pulls LLM models
just setup

# Start Go backend + Next.js frontend with hot reload
just dev
```

`just dev` starts Docker services, Ollama, the Go backend on `:8080`, and the Next.js frontend on `:3000`.

### Verifying Your Setup

```bash
just check    # Lint + typecheck (Go vet + Biome + tsc)
```

## Project Layout

```
internal/
├── agent/         # Chat agent: SDK wiring, session management, SSE streaming
├── app/           # Dependency wiring (wire.go)
├── config/        # Environment configuration, embedded SearXNG settings
├── events/        # SurrealDB event store for chat sessions
├── models/        # Shared data structures
├── orchestrator/  # Research pipeline: search → ingest → summarize
├── searcher/      # SearXNG HTTP client
├── server/        # Echo HTTP handlers, SSE writer, middleware
├── subprocess/    # Managed SurrealDB and SearXNG subprocesses (desktop mode)
└── tools/         # Agent tools: web_search, search_knowledge, store_knowledge

frontend/src/
├── app/           # Next.js pages, API client, hooks, types
├── components/    # React components (chat, research, graph, nav)
└── lib/           # Zustand stores and utilities

api/openapi.yaml   # OpenAPI 3.1 spec (source of truth for API types)
cmd/desktop/       # Wails desktop app entry point
```

### Key Dependencies

- **adk** (`github.com/urmzd/adk`): Agent SDK — typed deltas, provider interface, tool registry
- **kgdk** (`github.com/urmzd/kgdk`): Knowledge graph SDK — SurrealDB store, extraction, embeddings
- **SurrealDB**: Graph + document database (Docker in web mode, managed subprocess in desktop mode)
- **SearXNG**: Metasearch engine (Docker in web mode, managed subprocess in desktop mode)
- **Ollama**: Local LLM provider

## Making Changes

### Workflow

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run checks: `just check`
5. Submit a pull request

### API Changes

The API spec lives in `api/openapi.yaml`. When changing endpoints:

1. Update `api/openapi.yaml`
2. Update Go handlers in `internal/server/`
3. Regenerate frontend types: `just generate`

Do not edit files in `frontend/src/generated/` — they are auto-generated.

### Code Conventions

**Go:**

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep handlers thin — business logic belongs in `internal/agent/` or `internal/orchestrator/`
- Use the existing config pattern (`internal/config/`) for new environment variables
- Tools implement `core.Tool` from the adk SDK

**TypeScript:**

- Lint and format with [Biome](https://biomejs.dev): `npx biome check .`
- Follow the existing Next.js App Router patterns
- Use Zustand for client-side state
- Use shadcn/ui for UI primitives
- Use Streamdown for streaming markdown rendering

### Commit Messages

Use conventional commit prefixes:

- `feat:` — new feature
- `fix:` — bug fix
- `docs:` — documentation changes
- `refactor:` — code restructuring without behavior change
- `build:` — build system or dependency changes

## Desktop Mode

Zoro can run as a standalone desktop app via [Wails](https://wails.io/). In desktop mode, SurrealDB and SearXNG are managed as subprocesses (no Docker required).

```bash
just build-desktop    # Build the desktop binary
just dev-desktop      # Run in dev mode
```

## Reporting Issues

Open an issue on GitHub with:

- A clear description of the problem or suggestion
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Environment details (OS, Go version, Node version)

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
