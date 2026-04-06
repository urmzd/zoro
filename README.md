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

Zoro searches the web and local files, extracts entities and relationships using local LLMs, and stores everything locally — your data never leaves your machine.

## Features

- **Knowledge graph** — entities and relationships extracted from web content, stored in PostgreSQL + pgvector
- **Multi-turn chat** — conversational agent with session persistence
- **Local file exploration** — search and read local files to ground answers in your own data
- **Deep research** — automated pipeline: web search, file exploration, knowledge ingestion, LLM synthesis
- **Fully local** — Ollama for inference, SearXNG for search, PostgreSQL for storage
- **CLI interface** — chat, research, search, knowledge management, and graph visualization

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

## Configuration

All environment variables are optional. Defaults work out of the box.

| Variable | Default | Description |
|---|---|---|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODEL` | `gemma4:latest` | Main LLM for reasoning |
| `EMBEDDING_MODEL` | `nomic-embed-text` | Model used for embeddings |
| `POSTGRES_URL` | `postgres://zoro:zoro@localhost:5432/zoro?sslmode=disable` | PostgreSQL connection URL |
| `SEARXNG_URL` | (managed subprocess) | Set to use external SearXNG |
| `ZORO_DATA_DIR` | `~/.config/zoro` | App data directory |

When `SEARXNG_URL` is unset, Zoro manages SearXNG as a subprocess. When running via `just dev`, it defaults to `http://127.0.0.1:8888` (Docker).

## Agent Skill

This repo's conventions are available as portable agent skills in [`skills/`](skills/).

## Related

- [`saige`](https://github.com/urmzd/saige) — agent loop, knowledge graph, pgvector store, extraction pipeline, Ollama adapter

## License

[Apache License 2.0](LICENSE) &middot; [Third-party licenses](THIRD-PARTY-LICENSES.md)
