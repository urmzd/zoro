# RFC 001 — Neo4j Knowledge Graph Integration

| Field       | Value                          |
|-------------|--------------------------------|
| **Status**  | Draft                          |
| **Created** | 2026-03-01                     |
| **Scope**   | Backend, Frontend, Infra       |

---

## Table of Contents

1. [Motivation](#1-motivation)
2. [Neo4j Graph Schema (Ontology)](#2-neo4j-graph-schema-ontology)
3. [Relationship Types](#3-relationship-types)
4. [Entity Resolution / Deduplication](#4-entity-resolution--deduplication)
5. [Integration Architecture](#5-integration-architecture)
6. [Enrichment Flow](#6-enrichment-flow)
7. [Cross-Session Querying](#7-cross-session-querying)
8. [New API Endpoints](#8-new-api-endpoints)
9. [Frontend Changes](#9-frontend-changes)
10. [Configuration](#10-configuration)
11. [Migration Strategy](#11-migration-strategy)
12. [Sequence Diagram](#12-sequence-diagram)
13. [Risk Mitigation](#13-risk-mitigation)

---

## 1. Motivation

Zoro's research pipeline currently stores all data in an in-memory `dict[UUID, ResearchSession]` inside `ResearchOrchestrator` (`backend/src/zoro/services/orchestrator.py:22`). This design has five fundamental limitations:

### Ephemeral State

Sessions exist only in process memory. A server restart, deployment, or crash destroys all accumulated research. There is no persistence layer — the `sessions` dict is the only store.

### Session Isolation

Each call to `POST /research` creates a completely independent `ResearchSession`. Searching "quantum computing" and then "quantum entanglement" produces two disconnected graphs. There is no mechanism for the second session to discover that closely related topics were already explored in the first.

### Duplicate Entities

If two searches both discover a topic like "wave-particle duality", Zoro creates two separate `TopicSummary` objects with independent summaries, keywords, and embeddings. There is no deduplication — the same concept is represented N times across N sessions that encounter it.

### No Knowledge Accumulation

Research is stateless. The system never gets smarter. A user who has run 100 searches starts each new search with the same blank slate as a first-time user. There is no way to build on prior discoveries, refine understanding, or surface connections that span sessions.

### Intra-Session-Only Similarity

The `GraphBuilder` (`backend/src/zoro/services/graph.py`) computes cosine similarity between topic embeddings, but only within a single session's topic list. Cross-session similarity — the most valuable kind for knowledge discovery — is structurally impossible with the current architecture.

### What This RFC Proposes

A Neo4j-backed persistent ontological knowledge graph that:

- **Persists** all research artifacts across restarts
- **Connects** topics, concepts, and sources across sessions
- **Deduplicates** entities via embedding-based entity resolution
- **Accumulates** knowledge — each session enriches a shared graph
- **Enables cross-session querying** — "show me everything related to X" across all research

---

## 2. Neo4j Graph Schema (Ontology)

### Node Labels

#### ResearchSession

Represents a single research run initiated by the user.

| Property      | Type     | Description                              |
|---------------|----------|------------------------------------------|
| `id`          | STRING   | UUID, primary key                        |
| `query`       | STRING   | Original user query                      |
| `status`      | STRING   | `pending \| searching \| analyzing \| grouping \| graphing \| complete \| error` |
| `created_at`  | DATETIME | Session creation timestamp               |
| `completed_at`| DATETIME | Session completion timestamp (nullable)  |

#### Query

Represents a normalized search query. Multiple sessions may share the same query.

| Property        | Type     | Description                            |
|-----------------|----------|----------------------------------------|
| `text`          | STRING   | Normalized query text (lowercased, trimmed) |
| `first_seen_at` | DATETIME | First time this query was issued       |
| `run_count`     | INTEGER  | Number of sessions that used this query |

#### Topic

The central knowledge node. Represents a distinct topic discovered during research.

| Property       | Type       | Description                                          |
|----------------|------------|------------------------------------------------------|
| `id`           | STRING     | UUID, primary key                                    |
| `name`         | STRING     | Canonical topic name (title-cased, deduplicated)     |
| `summary`      | STRING     | Accumulated summary (merged across sessions)         |
| `keywords`     | LIST<STRING> | Union of all keywords across sessions              |
| `embedding`    | LIST<FLOAT>  | EMA-updated embedding vector (dimensionality matches `nomic-embed-text`) |
| `first_seen_at`| DATETIME   | When this topic was first discovered                 |
| `last_seen_at` | DATETIME   | Most recent session that encountered this topic      |
| `session_count`| INTEGER    | Number of sessions that discovered this topic        |

#### Concept

A fine-grained concept extracted from topic analysis. Concepts are leaf-level knowledge atoms.

| Property   | Type   | Description                     |
|------------|--------|---------------------------------|
| `name`     | STRING | Normalized concept name         |
| `first_seen_at` | DATETIME | First discovery timestamp |

#### Source

A web source (URL) referenced by one or more topics.

| Property   | Type     | Description                     |
|------------|----------|---------------------------------|
| `url`      | STRING   | Canonical URL, primary key      |
| `title`    | STRING   | Page title from search result   |
| `snippet`  | STRING   | Search result snippet           |
| `first_seen_at` | DATETIME | First time this URL appeared |

#### TopicGroup

A cluster of related topics, produced by `OllamaAnalyzer.group_topics()`.

| Property   | Type   | Description                          |
|------------|--------|--------------------------------------|
| `name`     | STRING | Group label assigned by the LLM      |
| `session_id` | STRING | Session that produced this grouping |

### Constraints

```cypher
CREATE CONSTRAINT research_session_id IF NOT EXISTS
  FOR (s:ResearchSession) REQUIRE s.id IS UNIQUE;

CREATE CONSTRAINT query_text IF NOT EXISTS
  FOR (q:Query) REQUIRE q.text IS UNIQUE;

CREATE CONSTRAINT topic_id IF NOT EXISTS
  FOR (t:Topic) REQUIRE t.id IS UNIQUE;

CREATE CONSTRAINT topic_name IF NOT EXISTS
  FOR (t:Topic) REQUIRE t.name IS UNIQUE;

CREATE CONSTRAINT concept_name IF NOT EXISTS
  FOR (c:Concept) REQUIRE c.name IS UNIQUE;

CREATE CONSTRAINT source_url IF NOT EXISTS
  FOR (s:Source) REQUIRE s.url IS UNIQUE;
```

### Indexes

```cypher
-- Full-text index for search-as-you-type
CREATE FULLTEXT INDEX topic_search IF NOT EXISTS
  FOR (t:Topic) ON EACH [t.name, t.summary];

-- Vector index for embedding similarity (cosine, 768 dimensions for nomic-embed-text)
CREATE VECTOR INDEX topic_embedding IF NOT EXISTS
  FOR (t:Topic) ON (t.embedding)
  OPTIONS {indexConfig: {
    `vector.dimensions`: 768,
    `vector.similarity_function`: 'cosine'
  }};

-- Composite index for temporal queries
CREATE INDEX topic_last_seen IF NOT EXISTS
  FOR (t:Topic) ON (t.last_seen_at);
```

---

## 3. Relationship Types

### INITIATED

`(ResearchSession)-[:INITIATED]->(Query)`

A session initiates a query. Many sessions may initiate the same normalized query.

| Property | Type     | Description          |
|----------|----------|----------------------|
| `at`     | DATETIME | When the session started |

### DISCOVERED_IN

`(Topic)-[:DISCOVERED_IN]->(ResearchSession)`

Records which session first or subsequently discovered a topic.

| Property     | Type    | Description                            |
|--------------|---------|----------------------------------------|
| `at`         | DATETIME | Discovery timestamp                   |
| `is_new`     | BOOLEAN | `true` if this session created the topic; `false` if it matched an existing one |

### ANSWERS

`(Topic)-[:ANSWERS]->(Query)`

A topic is relevant to (answers) a query. Built when a topic is discovered during a query's session.

| Property    | Type  | Description                        |
|-------------|-------|------------------------------------|
| `relevance` | FLOAT | Cosine similarity between topic embedding and query embedding |

### HAS_CONCEPT

`(Topic)-[:HAS_CONCEPT]->(Concept)`

A topic contains a fine-grained concept (extracted from keywords).

| Property | Type     | Description              |
|----------|----------|--------------------------|
| `at`     | DATETIME | When the link was created |

### SOURCED_FROM

`(Topic)-[:SOURCED_FROM]->(Source)`

A topic's information originates from a specific web source.

| Property | Type   | Description               |
|----------|--------|---------------------------|
| `rank`   | INTEGER | Position in search results |

### SIMILAR_TO

`(Topic)-[:SIMILAR_TO]->(Topic)`

Two topics are semantically similar (but not duplicates — duplicates are merged).

| Property     | Type  | Description                       |
|--------------|-------|-----------------------------------|
| `score`      | FLOAT | Cosine similarity score (0.5–0.85 range; ≥0.85 triggers merge) |
| `computed_at`| DATETIME | When the similarity was last computed |

### SUBTOPIC_OF

`(Topic)-[:SUBTOPIC_OF]->(Topic)`

Hierarchical relationship between a narrower and broader topic.

| Property | Type     | Description              |
|----------|----------|--------------------------|
| `at`     | DATETIME | When the link was created |

### GROUPED_IN

`(Topic)-[:GROUPED_IN]->(TopicGroup)`

A topic belongs to a cluster group within a session.

| Property | Type     | Description              |
|----------|----------|--------------------------|
| `at`     | DATETIME | When the grouping occurred |

### CONTAINS

`(TopicGroup)-[:CONTAINS]->(Topic)`

Inverse of GROUPED_IN for graph traversal convenience.

| Property | Type     | Description              |
|----------|----------|--------------------------|
| `at`     | DATETIME | When the grouping occurred |

### RELATED_TO

`(Topic)-[:RELATED_TO]->(Topic)`

General-purpose semantic relationship inferred by the LLM during analysis. Captures relationships that are neither similarity nor hierarchy.

| Property     | Type   | Description                        |
|--------------|--------|------------------------------------|
| `relation`   | STRING | Free-text description (e.g., "is a prerequisite for") |
| `confidence` | FLOAT  | LLM-assigned confidence (0.0–1.0) |

---

## 4. Entity Resolution / Deduplication

When a new `TopicSummary` arrives from `OllamaAnalyzer.summarize_result()`, the knowledge graph must decide: is this a **new** topic, or a **rediscovery** of an existing one?

### Two-Stage Gate Algorithm (Topics)

Topics use a two-stage resolution process because they have rich semantic content that requires fuzzy matching.

#### Stage 1 — Vector Search (Recall Gate)

```cypher
CALL db.index.vector.queryNodes('topic_embedding', 5, $embedding)
YIELD node, score
WHERE score >= $vector_threshold
RETURN node.id AS id, node.name AS name, score
```

- Query the vector index with the incoming topic's embedding
- Retrieve top-5 candidates with cosine similarity ≥ `vector_threshold` (default: **0.80**)
- This is a high-recall gate — it casts a wide net

#### Stage 2 — Name Similarity (Precision Gate)

For each candidate from Stage 1, compute normalized name similarity:

```python
from difflib import SequenceMatcher

def name_similarity(a: str, b: str) -> float:
    return SequenceMatcher(None, a.lower(), b.lower()).ratio()
```

- If any candidate has name similarity ≥ `name_threshold` (default: **0.75**), **merge** the incoming topic into that candidate
- If multiple candidates pass both gates, select the one with the highest combined score: `0.6 * vector_score + 0.4 * name_score`
- If no candidate passes both gates, **create** a new Topic node

#### Why Two Stages?

| Scenario                          | Vector Only | Name Only | Two-Stage |
|-----------------------------------|-------------|-----------|-----------|
| "Quantum Entanglement" vs "Quantum Entanglement" | Merge | Merge | Merge |
| "QE" vs "Quantum Entanglement"    | Merge (embeddings close) | Miss (names differ) | Miss (correct — abbreviations need human review) |
| "Machine Learning" vs "Deep Learning" | Merge (embeddings close) | Miss (names differ) | Miss (correct — they're related, not identical) |
| "Neural Networks" vs "Neural Nets" | Merge | Merge | Merge |

### Exact-Match Resolution (Query, Source, Concept)

These entity types have natural keys that allow deterministic deduplication:

| Entity  | Match Key         | Normalization           |
|---------|-------------------|-------------------------|
| Query   | `text`            | `text.strip().lower()`  |
| Source  | `url`             | URL canonicalization (strip tracking params, normalize scheme) |
| Concept | `name`            | `name.strip().lower()`  |

All three use Cypher `MERGE` with `ON CREATE SET` / `ON MATCH SET` clauses — no application-level lookup required.

---

## 5. Integration Architecture

### New Service: `services/knowledge_graph.py`

```
backend/src/zoro/services/
├── orchestrator.py     # Modified: calls KnowledgeGraphService after each phase
├── analyzer.py         # Modified: computes per-topic embeddings (not just batch)
├── searcher.py         # Unchanged
├── graph.py            # Unchanged
└── knowledge_graph.py  # NEW: Neo4j integration
```

#### `KnowledgeGraphService` Class

```python
class KnowledgeGraphService:
    """Persistent knowledge graph backed by Neo4j."""

    def __init__(self, config: Neo4jConfig) -> None:
        self.driver: AsyncDriver  # neo4j async driver
        self.config: Neo4jConfig

    async def start(self) -> None:
        """Initialize driver and ensure schema (constraints, indexes)."""

    async def close(self) -> None:
        """Close the driver connection pool."""

    # --- Entity Resolution ---
    async def resolve_topic(
        self, name: str, summary: str, keywords: list[str],
        embedding: list[float], source_url: str
    ) -> tuple[str, bool]:
        """Resolve a TopicSummary to a Topic node. Returns (topic_id, is_new)."""

    # --- Session Lifecycle ---
    async def create_session(self, session_id: str, query: str) -> None:
        """Create ResearchSession and Query nodes, link with INITIATED."""

    async def complete_session(self, session_id: str) -> None:
        """Mark session as complete, set completed_at."""

    # --- Ingestion ---
    async def ingest_topic(
        self, session_id: str, query: str, topic: "TopicSummary",
        embedding: list[float], source: "SearchResult"
    ) -> None:
        """Full ingestion: resolve topic, create relationships, update embeddings."""

    async def ingest_groups(
        self, session_id: str, groups: list["TopicGroup"]
    ) -> None:
        """Create TopicGroup nodes and GROUPED_IN/CONTAINS relationships."""

    async def compute_cross_session_similarities(
        self, topic_ids: list[str], threshold: float = 0.5
    ) -> list[tuple[str, str, float]]:
        """Compute SIMILAR_TO edges between newly ingested topics and existing graph."""

    # --- Querying ---
    async def search_topics(
        self, query: str, embedding: list[float], limit: int = 20
    ) -> list[dict]:
        """Vector + fulltext hybrid search across all topics."""

    async def get_node_neighborhood(
        self, node_id: str, depth: int = 1
    ) -> dict:
        """Return a node and its relationships up to N hops."""

    async def get_stats(self) -> dict:
        """Return graph statistics (node counts, edge counts, etc.)."""

    async def get_full_graph(self, limit: int = 200) -> dict:
        """Return the full knowledge graph (capped for visualization)."""

    async def get_topic_evolution(self, topic_id: str) -> list[dict]:
        """Return the timeline of sessions that discovered/enriched a topic."""

    async def get_bridges(
        self, query_a: str, query_b: str
    ) -> list[dict]:
        """Find topics that bridge two queries."""
```

### Additive Integration (Feature Flag)

The knowledge graph is **additive** — it does not replace any existing functionality. The orchestrator pipeline continues to work exactly as before; the KG service is called as a side-effect after each phase.

```python
# In config.py
class Neo4jConfig(BaseSettings):
    model_config = SettingsConfigDict(env_prefix="NEO4J_")
    enabled: bool = False  # Feature flag — off by default
    uri: str = "bolt://localhost:7687"
    username: str = "neo4j"
    password: str = "password"
    database: str = "neo4j"
    # ...
```

When `NEO4J_ENABLED=false` (default), the orchestrator skips all KG calls. Existing behavior is completely unchanged.

### Per-Topic Embedding Change

Currently, `OllamaAnalyzer.compute_embeddings()` is called once at the end with all topic names batched together. For the KG, embeddings must be available per-topic at ingestion time.

**Change:** After `summarize_result()` produces a `TopicSummary`, immediately compute its embedding. Store it on the `TopicSummary` model (new optional field: `embedding: list[float] | None`). The existing batch call in `GraphBuilder` remains for backward compatibility but can read from pre-computed values.

---

## 6. Enrichment Flow

When `resolve_topic()` determines that an incoming topic matches an existing Topic node, the existing node is **enriched** rather than replaced.

### Merge Cypher

```cypher
MATCH (t:Topic {id: $existing_topic_id})
SET
  // EMA update for embedding (alpha = 0.3 weights new data at 30%)
  t.embedding = [i IN range(0, size(t.embedding) - 1) |
    0.7 * t.embedding[i] + 0.3 * $new_embedding[i]
  ],

  // Append new summary content
  t.summary = t.summary + '\n\n---\n\n' + $new_summary,

  // Union keywords (deduplicated)
  t.keywords = apoc.coll.toSet(t.keywords + $new_keywords),

  // Update temporal metadata
  t.last_seen_at = datetime(),
  t.session_count = t.session_count + 1
```

### Why EMA for Embeddings?

An exponential moving average (alpha = 0.3) balances stability and adaptation:

- **Stability**: The existing embedding (representing multiple sessions of evidence) retains 70% weight, preventing a single noisy session from corrupting a well-established topic's position in vector space.
- **Adaptation**: New evidence still shifts the embedding, so topics evolve as understanding deepens.
- **No storage overhead**: Unlike storing all historical embeddings, EMA is a single vector.

### Summary Accumulation

Summaries are concatenated with a horizontal-rule separator. This preserves the full provenance trail — each research session's perspective is retained. Future enhancement: an LLM-powered summary consolidation pass.

### Keyword Union

Keywords are merged into a deduplicated set via `apoc.coll.toSet()`. This broadens the topic's keyword coverage with each new session.

---

## 7. Cross-Session Querying

The knowledge graph enables queries that are impossible with the current in-memory architecture.

### Everything Related to X

Find all topics, concepts, and sources connected to a given topic within 2 hops:

```cypher
MATCH (t:Topic {name: $topic_name})
CALL apoc.path.subgraphAll(t, {
  maxLevel: 2,
  relationshipFilter: "SIMILAR_TO|HAS_CONCEPT|SOURCED_FROM|RELATED_TO|SUBTOPIC_OF"
}) YIELD nodes, relationships
RETURN nodes, relationships
```

### Bridging Topics Between Queries

Find topics that connect two different research queries:

```cypher
MATCH (q1:Query {text: $query_a})<-[:ANSWERS]-(bridge:Topic)-[:ANSWERS]->(q2:Query {text: $query_b})
RETURN bridge.name AS topic,
       bridge.summary AS summary,
       bridge.session_count AS times_seen
ORDER BY bridge.session_count DESC
```

### Knowledge Evolution Timeline

Track how a topic has evolved across sessions:

```cypher
MATCH (t:Topic {id: $topic_id})-[d:DISCOVERED_IN]->(s:ResearchSession)
MATCH (s)-[:INITIATED]->(q:Query)
RETURN s.id AS session_id,
       q.text AS query,
       d.at AS discovered_at,
       d.is_new AS was_new_discovery
ORDER BY d.at ASC
```

### Most Connected Concepts

Find the most interconnected topics in the knowledge graph:

```cypher
MATCH (t:Topic)
WITH t,
  size((t)-[:SIMILAR_TO]-()) AS similar_count,
  size((t)-[:ANSWERS]-()) AS query_count,
  size((t)-[:HAS_CONCEPT]-()) AS concept_count
WITH t, similar_count + query_count + concept_count AS total_connections
RETURN t.name AS topic,
       total_connections,
       t.session_count AS sessions
ORDER BY total_connections DESC
LIMIT 20
```

---

## 8. New API Endpoints

All new endpoints are mounted under the `/knowledge` prefix. They are only available when `NEO4J_ENABLED=true`.

### `GET /knowledge/search`

Hybrid vector + fulltext search across all topics.

**Query Parameters:**

| Param  | Type   | Default | Description                    |
|--------|--------|---------|--------------------------------|
| `q`    | string | required | Search query text             |
| `limit`| int    | 20      | Max results                    |

**Response:**

```json
{
  "results": [
    {
      "id": "uuid",
      "name": "Quantum Entanglement",
      "summary": "...",
      "keywords": ["quantum", "entanglement", "bell-states"],
      "score": 0.92,
      "session_count": 3,
      "last_seen_at": "2026-03-01T12:00:00Z"
    }
  ]
}
```

### `GET /knowledge/node/{node_id}`

Return a node and its immediate neighborhood.

**Query Parameters:**

| Param  | Type | Default | Description            |
|--------|------|---------|------------------------|
| `depth`| int  | 1       | Traversal depth (1–3)  |

**Response:**

```json
{
  "node": { "id": "...", "label": "Topic", "properties": {} },
  "neighbors": [
    { "id": "...", "label": "Concept", "properties": {}, "relationship": "HAS_CONCEPT" }
  ]
}
```

### `GET /knowledge/stats`

Return aggregate statistics about the knowledge graph.

**Response:**

```json
{
  "total_sessions": 42,
  "total_queries": 35,
  "total_topics": 128,
  "total_concepts": 512,
  "total_sources": 310,
  "total_similarities": 89,
  "avg_topics_per_session": 8.2,
  "most_connected_topic": "Machine Learning"
}
```

### `GET /knowledge/graph`

Return the full knowledge graph for visualization (capped).

**Query Parameters:**

| Param  | Type | Default | Description                    |
|--------|------|---------|--------------------------------|
| `limit`| int  | 200     | Max nodes to return            |

**Response:**

```json
{
  "nodes": [
    { "id": "...", "label": "Topic", "properties": { "name": "...", "session_count": 3 } }
  ],
  "edges": [
    { "source": "...", "target": "...", "type": "SIMILAR_TO", "properties": { "score": 0.87 } }
  ]
}
```

### `GET /knowledge/topic/{topic_id}/evolution`

Return the timeline of how a topic has been discovered and enriched.

**Response:**

```json
{
  "topic": { "id": "...", "name": "...", "session_count": 5 },
  "timeline": [
    {
      "session_id": "uuid",
      "query": "quantum computing basics",
      "discovered_at": "2026-02-15T10:00:00Z",
      "was_new": true
    },
    {
      "session_id": "uuid",
      "query": "quantum entanglement explained",
      "discovered_at": "2026-02-20T14:30:00Z",
      "was_new": false
    }
  ]
}
```

### `GET /knowledge/bridges`

Find topics that bridge two research queries.

**Query Parameters:**

| Param    | Type   | Default  | Description         |
|----------|--------|----------|---------------------|
| `query_a`| string | required | First query text    |
| `query_b`| string | required | Second query text   |

**Response:**

```json
{
  "query_a": "quantum computing",
  "query_b": "cryptography",
  "bridges": [
    {
      "id": "uuid",
      "name": "Shor's Algorithm",
      "summary": "...",
      "session_count": 2
    }
  ]
}
```

---

## 9. Frontend Changes

### New Route: `/knowledge`

Add a new page at `frontend/src/app/knowledge/page.tsx` that serves as the persistent knowledge exploration interface.

### Navigation Bar

Update `frontend/src/app/layout.tsx` to include a top-level navigation bar:

| Link       | Route          | Description             |
|------------|----------------|-------------------------|
| Search     | `/`            | Existing search form    |
| Knowledge  | `/knowledge`   | Knowledge graph explorer |

### Knowledge Page Layout

```
+---------------------------------------------------------------+
|  [Search Bar]                              [Stats Panel]      |
+---------------------------------------------------------------+
|                                             |                 |
|                                             | Node Detail     |
|       Knowledge Graph Visualization         | Panel           |
|       (full interactive graph)              |                 |
|                                             | - Properties    |
|                                             | - Neighbors     |
|                                             | - Evolution     |
|                                             |                 |
+---------------------------------------------------------------+
```

### Components

#### Knowledge Graph Visualization

- Reuse `@xyflow/react` (already a dependency) for the interactive graph
- Fetch data from `GET /knowledge/graph`
- Color nodes by label type:
  - **Topic**: blue
  - **Query**: green
  - **Concept**: purple
  - **Source**: orange
  - **TopicGroup**: gray
- Edge styling by relationship type (solid for SIMILAR_TO, dashed for RELATED_TO, etc.)
- Click a node to populate the Node Detail Panel

#### Stats Panel

- Fetch from `GET /knowledge/stats`
- Display key metrics as cards: total topics, total sessions, total connections
- Show "Most Connected Topic" as a highlighted card

#### Node Detail Panel

- Triggered by clicking a node in the graph
- Fetch from `GET /knowledge/node/{id}?depth=1`
- Display:
  - Node properties (name, summary, keywords, session count)
  - List of neighbors grouped by relationship type
  - For Topic nodes: evolution timeline from `GET /knowledge/topic/{id}/evolution`

#### Search Integration

- Search bar at the top calls `GET /knowledge/search?q=...`
- Results highlight matching nodes in the graph visualization
- Clicking a search result centers the graph on that node

---

## 10. Configuration

### `Neo4jConfig`

Add to `backend/src/zoro/config.py`:

```python
class Neo4jConfig(BaseSettings):
    model_config = SettingsConfigDict(env_prefix="NEO4J_")

    # Feature flag
    enabled: bool = False

    # Connection
    uri: str = "bolt://localhost:7687"
    username: str = "neo4j"
    password: str = "password"
    database: str = "neo4j"

    # Connection pool
    max_connection_pool_size: int = 50
    connection_acquisition_timeout: float = 60.0  # seconds

    # Entity resolution thresholds
    vector_similarity_threshold: float = 0.80
    name_similarity_threshold: float = 0.75
    embedding_ema_alpha: float = 0.30

    # Cross-session similarity
    similarity_edge_threshold: float = 0.50
    max_similarity_candidates: int = 5
```

### Integration into `AppConfig`

```python
class AppConfig(BaseSettings):
    model_config = SettingsConfigDict(env_prefix="APP_")
    ollama: OllamaConfig = Field(default_factory=OllamaConfig)
    search: SearchConfig = Field(default_factory=SearchConfig)
    neo4j: Neo4jConfig = Field(default_factory=Neo4jConfig)  # NEW
    cors_origins: list[str] = ["http://localhost:3000"]
    host: str = "0.0.0.0"
    port: int = 8000
```

### Environment Variables

| Variable                          | Default              | Description                              |
|-----------------------------------|----------------------|------------------------------------------|
| `NEO4J_ENABLED`                   | `false`              | Feature flag for knowledge graph         |
| `NEO4J_URI`                       | `bolt://localhost:7687` | Neo4j Bolt protocol URI               |
| `NEO4J_USERNAME`                  | `neo4j`              | Neo4j authentication username            |
| `NEO4J_PASSWORD`                  | `password`           | Neo4j authentication password            |
| `NEO4J_DATABASE`                  | `neo4j`              | Neo4j database name                      |
| `NEO4J_MAX_CONNECTION_POOL_SIZE`  | `50`                 | Max connections in pool                  |
| `NEO4J_CONNECTION_ACQUISITION_TIMEOUT` | `60.0`          | Timeout for acquiring a connection (s)   |
| `NEO4J_VECTOR_SIMILARITY_THRESHOLD` | `0.80`            | Stage 1 entity resolution threshold      |
| `NEO4J_NAME_SIMILARITY_THRESHOLD` | `0.75`              | Stage 2 entity resolution threshold      |
| `NEO4J_EMBEDDING_EMA_ALPHA`       | `0.30`              | EMA weight for new embeddings            |
| `NEO4J_SIMILARITY_EDGE_THRESHOLD` | `0.50`              | Min score for SIMILAR_TO edges           |
| `NEO4J_MAX_SIMILARITY_CANDIDATES` | `5`                 | Top-K candidates for vector search       |

---

## 11. Migration Strategy

### Phase 1 — Backend KG Layer

**Goal:** Implement `KnowledgeGraphService` and integrate into the orchestrator pipeline.

**Steps:**

1. Add `neo4j` async driver to `pyproject.toml` dependencies
2. Create `Neo4jConfig` in `config.py`
3. Implement `services/knowledge_graph.py` with all methods
4. Add idempotent schema initialization (constraints + indexes) in `start()`
5. Modify `orchestrator.py` to call KG methods after each pipeline phase (guarded by `config.neo4j.enabled`)
6. Modify `analyzer.py` to compute per-topic embeddings
7. Add `embedding` field to `TopicSummary` model
8. Add unit tests with a test Neo4j instance
9. Add Docker Compose service for dev Neo4j

**Docker Compose (dev):**

```yaml
services:
  neo4j:
    image: neo4j:5-community
    ports:
      - "7474:7474"  # Browser
      - "7687:7687"  # Bolt
    environment:
      NEO4J_AUTH: neo4j/password
      NEO4J_PLUGINS: '["apoc", "graph-data-science"]'
    volumes:
      - neo4j_data:/data
      - neo4j_logs:/logs

volumes:
  neo4j_data:
  neo4j_logs:
```

**Schema Initialization (idempotent):**

The `KnowledgeGraphService.start()` method runs all `CREATE ... IF NOT EXISTS` constraint and index statements on startup. This is safe to run on every boot — Neo4j's `IF NOT EXISTS` clauses are no-ops when the schema already exists.

### Phase 2 — Frontend Knowledge View

**Goal:** Build the `/knowledge` page and visualization.

**Steps:**

1. Add knowledge API routes in `backend/src/zoro/routes/knowledge.py`
2. Mount routes conditionally when `NEO4J_ENABLED=true`
3. Add API client functions in `frontend/src/app/lib/api.ts`
4. Add TypeScript types for knowledge graph responses
5. Create `frontend/src/app/knowledge/page.tsx`
6. Create knowledge-specific components (graph viz, stats, detail panel)
7. Update `layout.tsx` with navigation bar
8. Add responsive layout and loading states

### Phase 3 — Enhanced Integration

**Goal:** Leverage the KG to improve research quality.

**Steps:**

1. Before searching, query the KG for existing topics related to the new query
2. Present "prior knowledge" in the research view (e.g., "3 related topics found from previous sessions")
3. Use KG-derived context to improve LLM prompts in `OllamaAnalyzer`
4. Add SSE events for KG ingestion progress (`kg_topic_resolved`, `kg_enriched`)
5. Cross-session similarity edges surfaced in the per-session graph view

---

## 12. Sequence Diagram

```
User            Frontend           Backend API        Orchestrator        KnowledgeGraphService     Neo4j
 |                 |                    |                  |                       |                   |
 |  POST /research |                    |                  |                       |                   |
 |---------------->|  POST /research    |                  |                       |                   |
 |                 |-- ---------------->|  create_session() |                       |                   |
 |                 |                    |----------------->|                       |                   |
 |                 |                    |                  |  kg.create_session()   |                   |
 |                 |                    |                  |---------------------->|  MERGE Session     |
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |  MERGE Query       |
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |  CREATE INITIATED  |
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |                   |
 |  GET /stream    |                    |                  |                       |                   |
 |---------------->|  EventSource       |                  |                       |                   |
 |                 |------------------->|  run_research()   |                       |                   |
 |                 |                    |----------------->|                       |                   |
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |  === SEARCH PHASE === |                   |
 |                 |                    |                  |  searcher.search()    |                   |
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |  === ANALYSIS PHASE ===                   |
 |                 |                    |                  |  For each SearchResult:                   |
 |                 |                    |                  |    analyzer.summarize_result()             |
 |                 |                    |                  |    analyzer.compute_embeddings([topic])    |
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |    kg.ingest_topic()  |                   |
 |                 |                    |                  |---------------------->|                   |
 |                 |                    |                  |                       |  resolve_topic()  |
 |                 |                    |                  |                       |  (Stage 1: vector)|
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |  (Stage 2: name)  |
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |                       |  IF new: CREATE   |
 |                 |                    |                  |                       |  IF existing: EMA |
 |                 |                    |                  |                       |  + summary append |
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |                       |  MERGE Source      |
 |                 |                    |                  |                       |  CREATE SOURCED_FROM
 |                 |                    |                  |                       |  CREATE ANSWERS    |
 |                 |                    |                  |                       |  CREATE DISCOVERED_IN
 |                 |                    |                  |                       |  MERGE Concepts    |
 |                 |                    |                  |                       |  CREATE HAS_CONCEPT|
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |                   |
 |  SSE: topic_discovered               |                  |                       |                   |
 |<----------------|<-------------------|<-----------------|                       |                   |
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |  === GROUPING PHASE ===                   |
 |                 |                    |                  |  analyzer.group_topics()                  |
 |                 |                    |                  |  kg.ingest_groups()   |                   |
 |                 |                    |                  |---------------------->|  MERGE TopicGroup |
 |                 |                    |                  |                       |  CREATE GROUPED_IN |
 |                 |                    |                  |                       |  CREATE CONTAINS   |
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |  === GRAPH PHASE ===  |                   |
 |                 |                    |                  |  graph.build()        |                   |
 |                 |                    |                  |  kg.compute_cross_    |                   |
 |                 |                    |                  |  session_similarities()|                  |
 |                 |                    |                  |---------------------->|  Vector search     |
 |                 |                    |                  |                       |  all existing      |
 |                 |                    |                  |                       |  topics            |
 |                 |                    |                  |                       |  CREATE SIMILAR_TO |
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |                   |
 |                 |                    |                  |  kg.complete_session()|                   |
 |                 |                    |                  |---------------------->|  SET completed_at  |
 |                 |                    |                  |                       |------------------>|
 |                 |                    |                  |                       |                   |
 |  SSE: research_complete              |                  |                       |                   |
 |<----------------|<-------------------|<-----------------|                       |                   |
```

---

## 13. Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| **Neo4j unavailable** | Medium | High — KG features fail | Feature flag (`NEO4J_ENABLED`) ensures core research pipeline is unaffected. KG calls are wrapped in try/except with logging; failures are non-fatal. Orchestrator continues without KG. |
| **Embedding model changes** | Low | High — vector index becomes inconsistent | Store `embedding_model` name on each Topic node. On model change, detect mismatch and trigger a re-embedding migration. Vector index rebuild is idempotent. |
| **False positive merges** | Medium | Medium — unrelated topics merged | Two-stage gate (vector + name) reduces false positives. Conservative default thresholds (0.80 vector, 0.75 name). Merges are logged with full provenance for manual review. Future: add an "unmerge" admin endpoint. |
| **False negative merges** | Low | Low — duplicate topics created | Duplicates are tolerable (they just add redundancy). Periodic batch deduplication job can clean up. Lower thresholds if false negatives are frequent. |
| **Neo4j query performance** | Medium | Medium — slow knowledge page | All queries use indexed lookups (constraints, vector index, fulltext index). `GET /knowledge/graph` caps results at 200 nodes. Connection pooling prevents connection exhaustion. Monitor query times via Neo4j query log. |
| **Added latency per research session** | Medium | Low — slightly slower research | KG ingestion is not on the critical path — it happens after each SSE event is already emitted. Research results stream to the user immediately; KG writes are fire-and-forget with async confirmation. If KG latency exceeds 500ms per topic, batch writes. |
| **Neo4j disk usage growth** | Low | Low — storage costs | Topic merging bounds growth (deduplicated). Source nodes are bounded by unique URLs. Session nodes are lightweight. Monitor with `GET /knowledge/stats`. Add TTL-based pruning for sessions older than N days if needed. |
| **APOC plugin unavailability** | Low | Medium — some queries fail | Only `apoc.coll.toSet` and `apoc.path.subgraphAll` are used. Provide pure-Cypher fallbacks: `UNWIND + COLLECT DISTINCT` for set union, variable-length path patterns for subgraph traversal. |
