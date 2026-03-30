# Zoro

Privacy-first research agent that builds a personal knowledge graph on your machine. Searches the web, extracts entities and relationships using local LLMs, and stores everything locally — your data never leaves your machine.

Zoro works two ways: as an [MCP](https://modelcontextprotocol.io/) server for use with Claude Code, Cursor, and other MCP clients, or as a standalone CLI for direct use from your terminal.

## Showcase

<table align="center">
  <tr>
    <td align="center"><strong>CLI Help</strong></td>
    <td align="center"><strong>Web Search</strong></td>
  </tr>
  <tr>
    <td align="center"><img src="showcase/cli-help.gif" alt="CLI help output" width="400"></td>
    <td align="center"><img src="showcase/cli-search.gif" alt="Web search from CLI" width="400"></td>
  </tr>
</table>

### Why no frontend?

Zoro is a personal research engine — it's yours and yours only. By exposing capabilities through MCP and CLI instead of a bundled UI, you use Zoro from whatever interface you already work in. Your knowledge graph follows you across tools rather than being locked behind a single app.

## Architecture

```
MCP client (Claude Code, Cursor, etc.)        Terminal
  └── zoro serve (MCP over stdio)              └── zoro <command>
        │                                            │
        └──────────────┬─────────────────────────────┘
                       │
                 Shared internals
                 ├── Chat agent (multi-turn, session persistence)
                 ├── Research pipeline (web search → ingest → LLM synthesis)
                 ├── Knowledge graph (PostgreSQL + pgvector)
                 ├── SearXNG (web search, Docker or managed subprocess)
                 └── Ollama (local LLMs)
```

**Key dependencies**:
- [`saige`](https://github.com/urmzd/saige) — agent loop, knowledge graph, pgvector store, extraction pipeline, Ollama adapter
- [`mcp-go`](https://github.com/mark3labs/mcp-go) — MCP server framework

## Prerequisites

- [Go](https://go.dev) 1.25+
- [Docker](https://docs.docker.com/get-docker/) with Docker Compose (for PostgreSQL + SearXNG)
- [Ollama](https://ollama.ai)
- [Just](https://github.com/casey/just)

## Quick Start

```bash
just setup   # pull Ollama models, fetch Go deps, start Docker services
just build   # build the binary
```

## Usage

### CLI

Zoro provides subcommands for direct terminal use:

```
zoro <command> [flags]

Commands:
  serve       Start MCP server on stdio (default)
  chat        Chat with Zoro
  research    Run deep research pipeline
  search      Search the web
  version     Print version
  help        Show help
```

**Chat** — multi-turn conversation with the agent:

```bash
zoro chat what are the latest advances in quantum computing

# Continue a previous session
zoro chat -s SESSION_ID tell me more about error correction
```

**Research** — deep research pipeline (web search, knowledge graph ingestion, LLM synthesis):

```bash
zoro research how do transformer attention mechanisms work
```

**Search** — web search via SearXNG:

```bash
zoro search rust async runtime comparison

# Machine-readable output
zoro search -json rust async runtime comparison
```

**Version**:

```bash
zoro version
```

### MCP Server

When run without a command (or with `serve`), Zoro starts an MCP server on stdio:

```bash
zoro          # or: zoro serve
```

Add to your MCP client config (e.g. `~/.claude.json` or project `.mcp.json`):

```json
{
  "mcpServers": {
    "zoro": {
      "command": "/path/to/zoro"
    }
  }
}
```

The MCP server exposes these tools:

| Tool | Description |
|------|-------------|
| `chat` | Multi-turn conversation with the agent. Pass `session_id` to continue a session. |
| `research` | Deep research pipeline: web search, knowledge graph ingestion, and LLM synthesis. |
| `web_search` | Search the web via SearXNG. |
| `search_knowledge` | Query the knowledge graph for stored facts and entities. |
| `store_knowledge` | Ingest text into the knowledge graph, extracting entities and relationships. |

## Development

```bash
just dev     # starts Docker services, Ollama, and runs the MCP server on stdio
```

### Justfile Commands

```
just setup            First-time setup (models + deps + Docker)
just dev              Run MCP server (starts Docker + Ollama if needed)
just build            Build the binary
just install          Install the binary to $GOPATH/bin
just stop             Stop Docker services
just check            Run go vet
just lint             Run golangci-lint
just vuln             Run govulncheck
just tidy             Tidy Go modules
just ci               Full CI gate (check + build)
just pull <model>     Download an Ollama model
just upgrade-ollama   Upgrade Ollama
just bench [model]    Benchmark model TPS (default: qwen3.5:4b)
```

### Docker Services

`docker-compose.yml` runs two services:

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| `postgres` | `pgvector/pgvector:pg17` | 5432 | Knowledge graph storage (pgvector) |
| `searxng` | `searxng/searxng:latest` | 8888 | Web search engine |

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
