# Contributing to Zoro

Thanks for your interest in contributing to Zoro. This guide covers how to set up a development environment and submit changes.

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Ollama](https://ollama.ai)
- [Just](https://github.com/casey/just)
- Python 3 (for test scripts)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/urmzd/zoro.git
cd zoro

# Start all services with hot reload
just dev
```

This starts Neo4j, pulls the default LLM model, starts Ollama, and runs both the API and frontend with hot reload.

If you prefer to run services individually:

```bash
just infra      # Neo4j
just serve      # Ollama
just api        # Go API (hot reload via air)
just web        # Next.js frontend (fast refresh)
```

### Verifying Your Setup

```bash
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
    ├── searcher.go      # Web search via headless Chrome
    └── ollama.go        # LLM integration

frontend/src/
├── app/         # Next.js pages, API client, types
├── components/  # React components
└── lib/         # State stores and utilities
```

## Making Changes

### Workflow

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run the end-to-end tests: `just test-e2e`
5. Submit a pull request

### Code Conventions

**Go (API):**

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep handlers thin — business logic belongs in `service/`
- Use the existing config pattern for new environment variables
- Update Swagger annotations when changing API endpoints: `just swagger`

**TypeScript (Frontend):**

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
just test-e2e           # Full pipeline test
just bench [model]      # LLM performance benchmarking
```

End-to-end tests validate the full pipeline: API health, research sessions, SSE streaming, and knowledge graph queries. Make sure Neo4j, Ollama, and the API are running before executing tests.

## Reporting Issues

Open an issue on GitHub with:

- A clear description of the problem or suggestion
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Environment details (OS, Go version, Node version)

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
