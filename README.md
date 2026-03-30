<p align="center">
  <h1 align="center">Zoro</h1>
  <p align="center">
    Privacy-first research agent that builds a personal knowledge graph on your machine.
    <br /><br />
    <a href="https://github.com/urmzd/zoro/releases">Download</a>
    &middot;
    <a href="https://github.com/urmzd/zoro/issues">Report Bug</a>
    &middot;
    <a href="https://pkg.go.dev/github.com/urmzd/zoro">Go Docs</a>
  </p>
</p>

<p align="center">
  <a href="https://github.com/urmzd/zoro/actions/workflows/ci.yml"><img src="https://github.com/urmzd/zoro/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

<table align="center" width="100%">
  <tr>
    <td align="center" width="50%"><img src="showcase/cli-help.gif" alt="CLI help" width="100%"></td>
    <td align="center" width="50%"><img src="showcase/cli-search.gif" alt="Web search" width="100%"></td>
  </tr>
</table>

Zoro searches the web, extracts entities and relationships using local LLMs, and stores everything locally — your data never leaves your machine. It works as an [MCP](https://modelcontextprotocol.io/) server for Claude Code, Cursor, and other MCP clients, or as a standalone CLI.

## Features

- **Knowledge graph** — entities and relationships extracted from web content, stored in PostgreSQL + pgvector
- **Multi-turn chat** — conversational agent with session persistence
- **Deep research** — automated pipeline: web search, knowledge ingestion, LLM synthesis
- **Fully local** — Ollama for inference, SearXNG for search, PostgreSQL for storage
- **Dual interface** — MCP server for tool-aware clients, CLI for direct terminal use

## Install

### From source

```bash
go install github.com/urmzd/zoro@latest
```

### From releases

Download a prebuilt binary from the [releases page](https://github.com/urmzd/zoro/releases).

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Docker Compose (for PostgreSQL + SearXNG)
- [Ollama](https://ollama.ai)
- [Just](https://github.com/casey/just) (for development)

## Quick Start

```bash
just setup   # pull Ollama models, fetch Go deps, start Docker services
just build   # build the binary
```

## Usage

### CLI

```bash
# Chat with Zoro
zoro chat what are the latest advances in quantum computing

# Continue a previous session
zoro chat -s SESSION_ID tell me more about error correction

# Deep research pipeline
zoro research how do transformer attention mechanisms work

# Web search
zoro search rust async runtime comparison
zoro search -json rust async runtime comparison   # machine-readable

# Knowledge graph visualization
zoro graph              # text format
zoro graph -format dot  # Graphviz DOT
zoro graph -format json # JSON

# Version
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

**MCP Tools:**

| Tool | Description |
|------|-------------|
| `chat` | Multi-turn conversation with the agent. Pass `session_id` to continue a session. |
| `research` | Deep research pipeline: web search, knowledge graph ingestion, and LLM synthesis. |
| `web_search` | Search the web via SearXNG. |
| `search_knowledge` | Query the knowledge graph for stored facts and entities. |
| `store_knowledge` | Ingest text into the knowledge graph, extracting entities and relationships. |
| `get_knowledge_graph` | Visualize the knowledge graph structure (text, DOT, or JSON format). |

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

## Agent Skill

This repo's conventions are available as portable agent skills in [`skills/`](skills/).

## Related

- [`saige`](https://github.com/urmzd/saige) — agent loop, knowledge graph, pgvector store, extraction pipeline, Ollama adapter
- [`go-sdk`](https://github.com/modelcontextprotocol/go-sdk) — MCP server framework

## License

[Apache License 2.0](LICENSE) &middot; [Third-party licenses](THIRD-PARTY-LICENSES.md)
