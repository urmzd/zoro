# Contributing to Zoro

Thanks for your interest in contributing to Zoro. This guide covers how to set up a development environment and submit changes.

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Ollama](https://ollama.ai)
- [Just](https://github.com/casey/just)
- [oag](https://github.com/urmzd/openapi-generator) (for API client generation)
- Python 3 (for test scripts)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/urmzd/zoro.git
cd zoro

# Start all services with hot reload
just dev
```

This starts Neo4j + SearXNG, pulls the default LLM model, starts Ollama, and runs both the API and frontend with hot reload.

If you prefer to run services individually:

```bash
just infra      # Neo4j + SearXNG
just serve      # Ollama
just api        # Go API (hot reload via air)
just web        # Next.js frontend (fast refresh)
```

### Verifying Your Setup

```bash
# Lint and typecheck
just check

# Run end-to-end tests
just test-e2e
```

## Project Layout

```
api/internal/
├── config/      # Environment configuration
├── handler/     # HTTP request handlers
├── model/       # Shared data structures
├── router/      # Route definitions and middleware
└── service/     # Core business logic
    ├── orchestrator.go  # Research pipeline coordination
    ├── knowledge.go     # Neo4j knowledge store operations
    ├── searcher.go      # Web search (SearXNG / ChromeDP)
    └── ollama.go        # LLM integration

frontend/src/
├── app/         # Next.js pages, API client, types
├── components/  # React components
├── generated/   # Auto-generated API client (do not edit)
└── lib/         # State stores and utilities
```

## Making Changes

### Workflow

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run checks: `just check`
5. Run the end-to-end tests: `just test-e2e`
6. Submit a pull request

### API Client Generation

The frontend API client is auto-generated from `openapi.yaml` using [oag](https://github.com/urmzd/openapi-generator). Do not edit files in `frontend/src/generated/` — they will be overwritten.

When you change API endpoints:

1. Update the Swagger annotations in the Go handlers
2. Regenerate Swagger docs: `just swagger`
3. Update `openapi.yaml` to match
4. Regenerate the client: `just generate`

### Code Conventions

**Go (API):**

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep handlers thin — business logic belongs in `service/`
- Use the existing config pattern for new environment variables
- Update Swagger annotations when changing API endpoints: `just swagger`

**TypeScript (Frontend):**

- Lint and format with [Biome](https://biomejs.dev): `npx biome check .`
- Follow the existing Next.js App Router patterns
- Use Zustand for client-side state
- Place reusable components under `src/components/`
- Use the shadcn/ui component library for UI primitives

### Commit Messages

Write clear, concise commit messages. Use conventional commit prefixes:

- `feat:` — new feature
- `fix:` — bug fix
- `docs:` — documentation changes
- `refactor:` — code restructuring without behavior change
- `build:` — build system or dependency changes
- `chore:` — maintenance tasks
- `test:` — test additions or changes

### Architecture Decisions

Significant changes should be proposed as an RFC in `docs/`. See `docs/rfc001-knowledge-graph.md` for an example. This helps maintain a record of design decisions and their rationale.

## Testing

```bash
just check              # Lint + typecheck
just test-e2e           # Full pipeline test
just bench [model]      # LLM performance benchmarking
```

End-to-end tests validate the full pipeline: API health, research sessions, SSE streaming, and knowledge graph queries. Make sure Neo4j, SearXNG, Ollama, and the API are running before executing tests.

## Reporting Issues

Open an issue on GitHub with:

- A clear description of the problem or suggestion
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Environment details (OS, Go version, Node version)

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
