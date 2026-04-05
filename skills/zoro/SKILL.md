---
name: zoro
description: >
  Privacy-first AI research agent with local knowledge graph and file exploration.
  Use when the user wants to search the web, run deep research, chat with an AI
  assistant, query or store knowledge in a local graph, visualize entity
  relationships, or search/read local files. Triggers on: research, knowledge
  graph, web search, file search, local files, SearXNG, Ollama, entity
  extraction, multi-turn chat, zoro.
user-invocable: true
allowed-tools: Bash, Read
---

# Zoro

Privacy-first research agent that builds a personal knowledge graph locally using Ollama, PostgreSQL+pgvector, and SearXNG. Also supports local file search and exploration.

## Prerequisites

Before running any command, verify services are up:

```bash
bash ${CLAUDE_SKILL_DIR}/scripts/preflight.sh
```

If preflight fails, start services with `just setup` from the project root.

## Commands

### Chat

Multi-turn conversational AI with tool use (web search, file search, knowledge graph).

```bash
# One-shot chat
zoro chat "What is Rust?"

# Continue a session
zoro chat -s SESSION_ID "Tell me more"

# JSON output (includes session_id for programmatic use)
zoro chat -json "What is Rust?"
```

**Session management:** Use `-json` to capture the `session_id` from the response, then pass it back with `-s` for continuity.

The chat agent has access to these tools internally:
- `web_search` — search the web for current information
- `file_search` — search local file contents by regex pattern
- `read_file` — read specific local files
- `search_knowledge` — query the knowledge graph
- `store_knowledge` — persist findings to the knowledge graph
- `get_knowledge_graph` — visualize entity relationships

### Research

Deep research pipeline: web search, knowledge ingestion, LLM synthesis. Returns a markdown summary.

```bash
zoro research "quantum computing applications"
```

### Web Search

Direct web search via SearXNG. Returns top results.

```bash
# Human-readable
zoro search "golang concurrency patterns"

# Machine-readable
zoro search -json "golang concurrency patterns"
```

### Knowledge Search

Query the knowledge graph for previously stored facts and entities.

```bash
# Basic search
zoro knowledge search "quantum computing"

# With options
zoro knowledge search -limit 20 -group "research-quantum" "entanglement"
```

### Knowledge Store

Ingest text into the knowledge graph. Extracts entities and relationships via Ollama.

```bash
# Store with default source
zoro knowledge store "Go is a statically typed language created at Google"

# Store with source label and group
zoro knowledge store -source "wikipedia" -group "research-go" "Go was designed by Robert Griesemer, Rob Pike, and Ken Thompson"
```

### Graph Visualization

Visualize the knowledge graph in multiple formats.

```bash
# Text summary (default)
zoro graph

# Graphviz DOT (pipe to dot for rendering)
zoro graph -format dot

# JSON (structured data)
zoro graph -format json

# Entity neighborhood
zoro graph -node ENTITY_UUID -depth 3
```

## Workflows

**Research then explore:**
```bash
zoro research "topic"
zoro knowledge search "specific aspect"
zoro graph -format text
```

**Iterative chat with persistence:**
```bash
# Start a session
zoro chat -json "initial question"
# Extract session_id from JSON, then:
zoro chat -s SESSION_ID "follow-up question"
```

**Manual knowledge building:**
```bash
zoro knowledge store -source "meeting notes" -group "project-x" "Key decisions from today..."
zoro knowledge search -group "project-x" "decisions"
```
