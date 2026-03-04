# Zoro

AI-powered research agent with a persistent knowledge graph. Zoro searches the web, extracts entities and relationships using local LLMs, and stores everything in a Neo4j graph database — so knowledge compounds across research sessions instead of disappearing.

## Why

Traditional research tools treat each session as a blank slate. Zoro builds a growing knowledge graph where entities, concepts, and relationships persist and connect across queries. Ask about quantum computing today, materials science tomorrow, and Zoro surfaces the connections automatically.

## Features

- **Web Research** — Automated Google search with headless Chrome, processing multiple results per query
- **Entity Extraction** — Local LLM extracts structured entities, relations, and facts from search results
- **Persistent Knowledge Graph** — Neo4j stores all research artifacts with vector + fulltext indexes
- **Cross-Session Discovery** — Entities deduplicated and linked across sessions via embedding similarity
- **Streaming Summaries** — Real-time LLM synthesis combining search results with existing graph knowledge
- **Interactive Graph Explorer** — Visual knowledge graph with search, filtering, and node inspection

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Frontend   │────▸│   Go API     │────▸│   Neo4j         │
│   Next.js    │◂────│   (Chi + SSE)│◂────│   Knowledge     │
│   React 19   │     │              │────▸│   Graph         │
└─────────────┘     │              │     └─────────────────┘
                    │              │────▸┌─────────────────┐
                    │              │◂────│   Ollama         │
                    │              │     │   Local LLM      │
                    └──────────────┘     └─────────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Google Search │
                    │ (ChromeDP)   │
                    └──────────────┘
```

## Tech Stack

| Layer          | Technology                                                  |
|----------------|-------------------------------------------------------------|
| Frontend       | Next.js 16, React 19, TypeScript, Tailwind CSS, Zustand    |
| Visualization  | React Flow, react-force-graph-2d                            |
| API            | Go 1.25, Chi v5, Server-Sent Events                        |
| Database       | Neo4j 5 Community (APOC, vector + fulltext indexes)         |
| LLM            | Ollama (Qwen 3.5:4b default, nomic-embed-text embeddings)  |
| Web Search     | ChromeDP (headless Chrome)                                  |
| Infrastructure | Docker Compose, Justfile                                    |

## Quick Start

**Prerequisites:** Docker, Go 1.25+, Node.js 18+, [Ollama](https://ollama.ai), [Just](https://github.com/casey/just)

```bash
# Start Neo4j
just infra

# Pull the default LLM model
just pull

# Start Ollama (if not already running)
just serve

# Run API + frontend with hot reload
just run
```

Or bring up everything at once:

```bash
just dev
```

The frontend is available at `http://localhost:3000` and the API at `http://localhost:8080`.

## Project Structure

```
├── api/
│   ├── cmd/zoro/           # Entry point
│   └── internal/
│       ├── config/         # Environment configuration
│       ├── handler/        # HTTP handlers
│       ├── model/          # Data structures
│       ├── router/         # Route definitions
│       └── service/        # Business logic (orchestrator, knowledge, search, LLM)
├── frontend/
│   └── src/
│       ├── app/            # Next.js pages and API client
│       ├── components/     # React components (graph, research, UI)
│       └── lib/            # Stores and utilities
├── docs/                   # RFCs and architecture docs
├── scripts/                # Benchmarking and test scripts
├── docker-compose.yml
└── justfile
```

## Configuration

All configuration is via environment variables with sensible defaults:

| Variable           | Default                    | Description                  |
|--------------------|----------------------------|------------------------------|
| `NEO4J_URI`        | `bolt://localhost:7687`    | Neo4j connection URI         |
| `NEO4J_USER`       | `neo4j`                    | Neo4j username               |
| `NEO4J_PASSWORD`   | `zoro_dev_password`        | Neo4j password               |
| `OLLAMA_HOST`      | `http://localhost:11434`   | Ollama server URL            |
| `OLLAMA_MODEL`     | `qwen3.5:4b`              | LLM model for extraction     |
| `EMBEDDING_MODEL`  | `nomic-embed-text`         | Model for vector embeddings  |
| `CORS_ORIGINS`     | `http://localhost:3000`    | Allowed CORS origins         |

## API Endpoints

| Method | Path                        | Description                          |
|--------|-----------------------------|--------------------------------------|
| POST   | `/api/research`             | Start a new research session         |
| GET    | `/api/research/{id}`        | Get session state                    |
| GET    | `/api/research/{id}/stream` | SSE event stream for live results    |
| GET    | `/api/knowledge/search`     | Hybrid vector + fulltext search      |
| GET    | `/api/knowledge/graph`      | Full graph for visualization         |
| GET    | `/api/knowledge/node/{id}`  | Node details with neighbors          |

## Available Commands

```bash
just infra          # Start Neo4j
just infra-down     # Stop infrastructure
just api            # Run API with hot reload
just web            # Run frontend dev server
just run            # Run API + frontend
just dev            # Run everything
just pull [model]   # Download Ollama model
just serve          # Start Ollama server
just swagger        # Generate API docs
just bench [model]  # Benchmark model performance
just test-e2e       # Run end-to-end tests
```

## License

[Apache License 2.0](LICENSE)
