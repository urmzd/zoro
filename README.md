# Zoro

Privacy-first research agent that builds a personal knowledge graph on your machine. Searches the web, extracts entities and relationships using local LLMs, and stores everything locally — your data never leaves your machine.

Zoro runs as an [MCP](https://modelcontextprotocol.io/) server over stdio, designed for use with Claude Code and other MCP clients.

## Architecture

```
MCP client (Claude Code, etc.)
  └── Zoro MCP server (stdio)
        ├── Chat agent (multi-turn, session persistence)
        ├── Research pipeline (web search → ingest → LLM synthesis)
        ├── Knowledge graph (PostgreSQL + pgvector)
        ├── SearXNG (web search, Docker or managed subprocess)
        └── Ollama (local LLMs — must be installed separately)
```

**SDKs**:
- `github.com/urmzd/saige` — agent loop, knowledge graph, PostgreSQL store (pgvector), extraction pipeline, embedder, Ollama adapter
- `github.com/mark3labs/mcp-go` — MCP server framework

## MCP Tools

| Tool | Description |
|------|-------------|
| `chat` | Multi-turn conversation with the agent. Pass `session_id` to continue a session. |
| `research` | Deep research pipeline: web search, knowledge graph ingestion, and LLM synthesis. |
| `web_search` | Search the web via SearXNG. |
| `search_knowledge` | Query the knowledge graph for stored facts and entities. |
| `store_knowledge` | Ingest text into the knowledge graph, extracting entities and relationships. |

## Prerequisites

- [Go](https://go.dev) 1.25+
- [Docker](https://docs.docker.com/get-docker/) with Docker Compose (for PostgreSQL + SearXNG)
- [Ollama](https://ollama.ai)
- [Just](https://github.com/casey/just)

## Quick Start

```bash
just setup   # install deps, start Docker services, pull Ollama models
just build   # build the binary
```

### Claude Code

Add to your MCP config (`~/.claude/claude_desktop_config.json` or project `.mcp.json`):

```json
{
  "mcpServers": {
    "zoro": {
      "command": "/path/to/zoro"
    }
  }
}
```

### Development

```bash
just dev   # starts Docker services, Ollama, and runs the MCP server on stdio
```

## Commands

```
just setup       First-time setup (deps + Docker + models)
just dev         Run MCP server (starts Docker + Ollama if needed)
just build       Build the MCP server binary
just install     Install the binary to $GOPATH/bin
just stop        Stop Docker services
just check       Run go vet
just ci          Full CI gate (check + build)
just pull <m>    Download an Ollama model
```

## Configuration

All environment variables are optional. Defaults work out of the box.

| Variable | Default | Description |
|---|---|---|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODEL` | `qwen3.5:4b` | Main LLM for reasoning |
| `OLLAMA_FAST_MODEL` | `qwen3.5:0.8b` | Fast model for lightweight tasks |
| `EMBEDDING_MODEL` | `nomic-embed-text` | Model used for embeddings |
| `POSTGRES_URL` | `postgres://zoro:zoro@localhost:5432/zoro?sslmode=disable` | PostgreSQL connection URL |
| `SEARXNG_URL` | (managed subprocess) | Set to use external SearXNG |
| `ZORO_DATA_DIR` | `~/.config/zoro` | App data directory |

When `SEARXNG_URL` is unset, Zoro manages SearXNG as a subprocess. When running via `just dev`, it defaults to `http://127.0.0.1:8888` (Docker).

## Third-Party Licenses

See [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md) for attribution of bundled dependencies (including SearXNG, AGPL-3.0).

## License

[Apache License 2.0](LICENSE)
