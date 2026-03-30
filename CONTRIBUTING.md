# Contributing to Zoro

Thanks for your interest in contributing to Zoro. This guide covers how to set up a development environment and submit changes.

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Ollama](https://ollama.ai)
- [Just](https://github.com/casey/just)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/urmzd/zoro.git
cd zoro

# First-time setup: installs deps, starts Docker (PostgreSQL + SearXNG), pulls LLM models
just setup

# Start MCP server with Docker services
just dev
```

### Verifying Your Setup

```bash
just ci    # Runs vet + tests + build
```

## Project Layout

```
internal/
├── agent/         # Chat agent: SDK wiring, session management
├── app/           # Dependency wiring (wire.go)
├── config/        # Environment configuration, embedded SearXNG settings
├── events/        # PostgreSQL event store for chat sessions
├── graph/         # Knowledge graph formatting (DOT, text)
├── mcp/           # MCP server setup + tool handlers
├── models/        # Shared data structures
├── orchestrator/  # Research pipeline: search → ingest → summarize
├── searcher/      # SearXNG HTTP client
├── subprocess/    # Managed SearXNG subprocess
└── tools/         # Agent tools: web_search, search_knowledge, store_knowledge, get_knowledge_graph
```

### Key Dependencies

- **saige** (`github.com/urmzd/saige`): Agent SDK — agent loop, knowledge graph, pgvector store, extraction, embeddings, Ollama adapter
- **go-sdk** (`github.com/modelcontextprotocol/go-sdk`): MCP server framework
- **PostgreSQL**: Relational database with pgvector extension (Docker)
- **SearXNG**: Metasearch engine (Docker or managed subprocess)
- **Ollama**: Local LLM provider

## Making Changes

### Workflow

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run checks: `just ci`
5. Submit a pull request

### Testing

Tests live alongside source files as `*_test.go`. Run with:

```bash
just test    # go test -race -count=1 ./...
```

Write tests for new functionality. Test files exist for config, graph formatting, searcher, tools, and MCP handlers. Use `httptest.NewServer` for HTTP-dependent tests and mock implementations of `kgtypes.Graph` for knowledge graph tests.

### Code Conventions

**Go:**

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep handlers thin — business logic belongs in `internal/agent/` or `internal/orchestrator/`
- Use the existing config pattern (`internal/config/`) for new environment variables
- Tools implement `saige/agent/types.Tool`: `Definition()` + `Execute()`
- MCP handlers use `github.com/modelcontextprotocol/go-sdk/mcp`

### Commit Messages

Use conventional commit prefixes:

- `feat:` — new feature
- `fix:` — bug fix
- `docs:` — documentation changes
- `refactor:` — code restructuring without behavior change
- `build:` — build system or dependency changes
- `test:` — test additions or changes

## Reporting Issues

Open an issue on GitHub with:

- A clear description of the problem or suggestion
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Environment details (OS, Go version)

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
