# Zoro

Privacy-first research agent that builds a personal knowledge graph on your machine. Searches the web, extracts entities and relationships using local LLMs, and stores everything locally — your data never leaves your machine.

## Quick Start

**Prerequisites:** [Ollama](https://ollama.ai), [Rust](https://rustup.rs), Node.js 20+, [Just](https://github.com/casey/just), Python 3.10+ (build-time, for SearXNG sidecar)

```bash
just setup   # install deps + build SearXNG + pull models (one-time)
just dev     # launch the app
```

That's it. Ollama starts automatically if it isn't already running. SearXNG is bundled as a sidecar and starts/stops with the app — no Docker needed.

## Build

```bash
just build
```

Produces a native installer for your platform (`.dmg` / `.msi` / `.AppImage`).

## How It Works

```
Tauri Desktop App
├── Rust Backend
│   ├── Embedded SurrealDB (RocksDB — no separate server)
│   ├── Ollama client (streaming chat + embeddings)
│   ├── SearXNG sidecar (bundled, auto-managed)
│   ├── Agent loop (chat with tool use)
│   └── Tools: web_search, search_knowledge, store_knowledge
└── React Frontend (webview)
    ├── Chat with streaming markdown
    ├── Knowledge graph visualization
    └── Research explorer
```

External dependency: Ollama only. SearXNG is bundled as a sidecar binary.

## Configuration

Environment variables (all optional — defaults work out of the box):

| Variable           | Default                  | Description               |
|--------------------|--------------------------|---------------------------|
| `OLLAMA_HOST`      | `http://localhost:11434` | Ollama server URL         |
| `OLLAMA_MODEL`     | `qwen3.5:4b`            | Main LLM model            |
| `OLLAMA_FAST_MODEL`| `qwen3.5:0.8b`          | Fast model (autocomplete) |
| `EMBEDDING_MODEL`  | `nomic-embed-text`       | Embedding model           |

Data lives in your OS data directory (`~/Library/Application Support/zoro/` on macOS).

## Commands

```
just setup           # First-time setup (includes SearXNG build)
just dev             # Run in dev mode (SearXNG starts automatically)
just build-searxng   # Rebuild SearXNG sidecar binary
just build           # Production build
just check           # Lint + test everything
just pull <model>    # Download an Ollama model
just upgrade-ollama  # Upgrade Ollama
```

## Third-Party Licenses

See [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md) for attribution of bundled dependencies (including SearXNG, AGPL-3.0).

## License

[Apache License 2.0](LICENSE)
