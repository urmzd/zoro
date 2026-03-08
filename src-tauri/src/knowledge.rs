use anyhow::{Context, Result};
use surrealdb::engine::local::RocksDb;
use surrealdb::Surreal;

use crate::models::*;
use crate::ollama::OllamaClient;

pub struct KnowledgeStore {
    db: Surreal<surrealdb::engine::local::Db>,
}

impl KnowledgeStore {
    pub async fn new(path: &std::path::Path) -> Result<Self> {
        if let Some(parent) = path.parent() {
            std::fs::create_dir_all(parent)?;
        }
        let db = Surreal::new::<RocksDb>(path.to_str().unwrap_or("zoro.db"))
            .await
            .context("open surrealdb")?;
        db.use_ns("zoro").use_db("zoro").await?;
        Ok(Self { db })
    }

    pub async fn ensure_schema(&self) -> Result<()> {
        let statements = [
            "DEFINE TABLE IF NOT EXISTS entity SCHEMAFULL",
            "DEFINE FIELD IF NOT EXISTS uuid ON entity TYPE string",
            "DEFINE FIELD IF NOT EXISTS name ON entity TYPE string",
            "DEFINE FIELD IF NOT EXISTS type ON entity TYPE string",
            "DEFINE FIELD IF NOT EXISTS summary ON entity TYPE string",
            "DEFINE FIELD IF NOT EXISTS embedding ON entity TYPE option<array<float>>",
            "DEFINE INDEX IF NOT EXISTS entity_uuid ON entity FIELDS uuid UNIQUE",
            "DEFINE INDEX IF NOT EXISTS entity_name_type ON entity FIELDS name, type UNIQUE",
            "DEFINE INDEX IF NOT EXISTS entity_embedding ON entity FIELDS embedding MTREE DIMENSION 768 DIST COSINE",
            "DEFINE ANALYZER IF NOT EXISTS entity_analyzer TOKENIZERS blank, class FILTERS snowball(english)",
            "DEFINE INDEX IF NOT EXISTS entity_fulltext ON entity FIELDS name, summary SEARCH ANALYZER entity_analyzer BM25",
            "DEFINE TABLE IF NOT EXISTS episode SCHEMAFULL",
            "DEFINE FIELD IF NOT EXISTS uuid ON episode TYPE string",
            "DEFINE FIELD IF NOT EXISTS name ON episode TYPE string",
            "DEFINE FIELD IF NOT EXISTS body ON episode TYPE string",
            "DEFINE FIELD IF NOT EXISTS source ON episode TYPE string",
            "DEFINE FIELD IF NOT EXISTS group_id ON episode TYPE string",
            "DEFINE FIELD IF NOT EXISTS created_at ON episode TYPE datetime DEFAULT time::now()",
            "DEFINE INDEX IF NOT EXISTS episode_uuid ON episode FIELDS uuid UNIQUE",
            "DEFINE TABLE IF NOT EXISTS relates_to SCHEMAFULL TYPE RELATION IN entity OUT entity",
            "DEFINE FIELD IF NOT EXISTS uuid ON relates_to TYPE string",
            "DEFINE FIELD IF NOT EXISTS type ON relates_to TYPE string",
            "DEFINE FIELD IF NOT EXISTS fact ON relates_to TYPE string",
            "DEFINE INDEX IF NOT EXISTS relates_to_uuid ON relates_to FIELDS uuid UNIQUE",
            "DEFINE TABLE IF NOT EXISTS mentions SCHEMAFULL TYPE RELATION IN episode OUT entity",
        ];

        for stmt in &statements {
            if let Err(e) = self.db.query(*stmt).await {
                log::warn!("schema statement warning: {e} (statement: {stmt})");
            }
        }
        Ok(())
    }

    pub async fn add_episode(
        &self,
        req: &EpisodeRequest,
        ollama: &OllamaClient,
    ) -> Result<EpisodeResponse> {
        let (entities, relations) = ollama.extract_entities(&req.body).await?;

        let mut response_entities = Vec::new();
        let mut entity_uuids: std::collections::HashMap<String, String> =
            std::collections::HashMap::new();
        let mut entity_record_ids: std::collections::HashMap<String, String> =
            std::collections::HashMap::new();

        for e in &entities {
            let ent_uuid = uuid::Uuid::new_v4().to_string();

            let embedding = match ollama.embed(&format!("{} {}", e.name, e.summary)).await {
                Ok(emb) => Some(emb),
                Err(err) => {
                    log::warn!("embedding error for entity {}: {err}", e.name);
                    None
                }
            };

            let existing: Vec<EntityRecord> = self
                .db
                .query("SELECT id, uuid FROM entity WHERE name = $name AND type = $type LIMIT 1")
                .bind(("name", e.name.clone()))
                .bind(("type", e.entity_type.clone()))
                .await?
                .take(0)
                .unwrap_or_default();

            let (final_uuid, record_id) = if let Some(existing) = existing.first() {
                let rid = existing.id.to_string();
                if let Some(ref emb) = embedding {
                    let _ = self
                        .db
                        .query("UPDATE type::thing($id) SET summary = $summary, embedding = $embedding")
                        .bind(("id", rid.clone()))
                        .bind(("summary", e.summary.clone()))
                        .bind(("embedding", emb.clone()))
                        .await;
                } else {
                    let _ = self
                        .db
                        .query("UPDATE type::thing($id) SET summary = $summary")
                        .bind(("id", rid.clone()))
                        .bind(("summary", e.summary.clone()))
                        .await;
                }
                (existing.uuid.clone(), rid)
            } else {
                let create_result: Vec<EntityRecord> = if let Some(ref emb) = embedding {
                    self.db
                        .query("CREATE entity SET uuid = $uuid, name = $name, type = $type, summary = $summary, embedding = $embedding RETURN id, uuid")
                        .bind(("uuid", ent_uuid.clone()))
                        .bind(("name", e.name.clone()))
                        .bind(("type", e.entity_type.clone()))
                        .bind(("summary", e.summary.clone()))
                        .bind(("embedding", emb.clone()))
                        .await?
                        .take(0)
                        .unwrap_or_default()
                } else {
                    self.db
                        .query("CREATE entity SET uuid = $uuid, name = $name, type = $type, summary = $summary RETURN id, uuid")
                        .bind(("uuid", ent_uuid.clone()))
                        .bind(("name", e.name.clone()))
                        .bind(("type", e.entity_type.clone()))
                        .bind(("summary", e.summary.clone()))
                        .await?
                        .take(0)
                        .unwrap_or_default()
                };

                if let Some(c) = create_result.first() {
                    (ent_uuid.clone(), c.id.to_string())
                } else {
                    continue;
                }
            };

            entity_uuids.insert(e.name.clone(), final_uuid.clone());
            entity_record_ids.insert(final_uuid.clone(), record_id);
            response_entities.push(Entity {
                uuid: final_uuid,
                name: e.name.clone(),
                entity_type: e.entity_type.clone(),
                summary: e.summary.clone(),
            });
        }

        let mut response_relations = Vec::new();
        for r in &relations {
            let src_uuid = match entity_uuids.get(&r.source) {
                Some(u) => u.clone(),
                None => continue,
            };
            let tgt_uuid = match entity_uuids.get(&r.target) {
                Some(u) => u.clone(),
                None => continue,
            };
            let src_rid = match entity_record_ids.get(&src_uuid) {
                Some(r) => r.clone(),
                None => continue,
            };
            let tgt_rid = match entity_record_ids.get(&tgt_uuid) {
                Some(r) => r.clone(),
                None => continue,
            };

            let rel_uuid = uuid::Uuid::new_v4().to_string();

            if let Err(err) = self
                .db
                .query("RELATE type::thing($src)->relates_to->type::thing($tgt) SET uuid = $uuid, type = $type, fact = $fact")
                .bind(("src", src_rid))
                .bind(("tgt", tgt_rid))
                .bind(("uuid", rel_uuid.clone()))
                .bind(("type", r.relation_type.clone()))
                .bind(("fact", r.fact.clone()))
                .await
            {
                log::warn!("create relation error: {err}");
                continue;
            }

            response_relations.push(Relation {
                uuid: rel_uuid,
                source_uuid: src_uuid,
                target_uuid: tgt_uuid,
                relation_type: r.relation_type.clone(),
                fact: r.fact.clone(),
            });
        }

        let episode_uuid = uuid::Uuid::new_v4().to_string();

        let ep_results: Vec<EpisodeRecord> = self
            .db
            .query("CREATE episode SET uuid = $uuid, name = $name, body = $body, source = $source, group_id = $group_id RETURN id")
            .bind(("uuid", episode_uuid.clone()))
            .bind(("name", req.name.clone()))
            .bind(("body", req.body.clone()))
            .bind(("source", req.source.clone()))
            .bind(("group_id", req.group_id.clone()))
            .await?
            .take(0)
            .unwrap_or_default();

        if let Some(ep) = ep_results.first() {
            let ep_id = ep.id.to_string();
            for ent_uuid in entity_uuids.values() {
                if let Some(ent_rid) = entity_record_ids.get(ent_uuid) {
                    let _ = self
                        .db
                        .query("RELATE type::thing($ep)->mentions->type::thing($ent)")
                        .bind(("ep", ep_id.clone()))
                        .bind(("ent", ent_rid.clone()))
                        .await;
                }
            }
        }

        Ok(EpisodeResponse {
            uuid: episode_uuid,
            name: req.name.clone(),
            entity_nodes: response_entities,
            episodic_edges: response_relations,
        })
    }

    pub async fn search_facts(
        &self,
        query: &str,
        group_id: &str,
        ollama: &OllamaClient,
    ) -> Result<SearchFactsResponse> {
        let embedding = match ollama.embed(query).await {
            Ok(emb) => emb,
            Err(_) => return Ok(SearchFactsResponse { facts: vec![] }),
        };

        let rows: Vec<FactRow> = if !group_id.is_empty() {
            self.db
                .query(
                    r#"SELECT
                        r.uuid AS r_uuid, r.type AS r_type, r.fact AS r_fact,
                        node.uuid AS src_uuid, node.name AS src_name, node.type AS src_type, node.summary AS src_summary,
                        other.uuid AS tgt_uuid, other.name AS tgt_name, other.type AS tgt_type, other.summary AS tgt_summary,
                        vector::similarity::cosine(node.embedding, $emb) AS score
                    FROM entity AS node
                    WHERE node.embedding <|20|> $emb
                    AND (SELECT VALUE id FROM episode WHERE group_id = $group_id AND ->mentions->entity CONTAINS node.id) != []
                    SPLIT r
                    LET r = (SELECT * FROM relates_to WHERE in = node.id OR out = node.id)
                    LET other = IF r.in = node.id THEN (SELECT * FROM type::thing(r.out)) ELSE (SELECT * FROM type::thing(r.in)) END
                    ORDER BY score DESC"#,
                )
                .bind(("emb", embedding))
                .bind(("group_id", group_id.to_string()))
                .await?
                .take(0)
                .unwrap_or_default()
        } else {
            self.db
                .query(
                    r#"LET $matches = (SELECT id, uuid, name, type, summary, embedding,
                        vector::similarity::cosine(embedding, $emb) AS score
                        FROM entity WHERE embedding <|20|> $emb ORDER BY score DESC);
                    SELECT
                        r.uuid AS r_uuid, r.type AS r_type, r.fact AS r_fact,
                        m.uuid AS src_uuid, m.name AS src_name, m.type AS src_type, m.summary AS src_summary,
                        other.uuid AS tgt_uuid, other.name AS tgt_name, other.type AS tgt_type, other.summary AS tgt_summary,
                        m.score AS score
                    FROM $matches AS m,
                        (SELECT * FROM relates_to WHERE in = m.id OR out = m.id) AS r,
                        IF r.in = m.id THEN (SELECT * FROM entity WHERE id = r.out)[0] ELSE (SELECT * FROM entity WHERE id = r.in)[0] END AS other
                    ORDER BY score DESC"#,
                )
                .bind(("emb", embedding))
                .await?
                .take(1)
                .unwrap_or_default()
        };

        let mut facts = Vec::new();
        let mut seen = std::collections::HashSet::new();
        for row in rows {
            if row.r_uuid.is_empty() || seen.contains(&row.r_uuid) {
                continue;
            }
            seen.insert(row.r_uuid.clone());
            facts.push(Fact {
                uuid: row.r_uuid,
                name: row.r_type,
                fact: row.r_fact,
                source_node: Entity {
                    uuid: row.src_uuid,
                    name: row.src_name,
                    entity_type: row.src_type,
                    summary: row.src_summary,
                },
                target_node: Entity {
                    uuid: row.tgt_uuid,
                    name: row.tgt_name,
                    entity_type: row.tgt_type,
                    summary: row.tgt_summary,
                },
            });
        }

        Ok(SearchFactsResponse { facts })
    }

    pub async fn get_graph(&self, limit: i64) -> Result<GraphData> {
        let rows: Vec<GraphRow> = self
            .db
            .query(
                r#"SELECT
                    a.uuid AS a_uuid, a.name AS a_name, a.type AS a_type, a.summary AS a_summary,
                    r.uuid AS r_uuid, r.type AS r_type, r.fact AS r_fact,
                    b.uuid AS b_uuid, b.name AS b_name, b.type AS b_type, b.summary AS b_summary
                FROM relates_to AS r
                LET a = (SELECT * FROM entity WHERE id = r.in)[0]
                LET b = (SELECT * FROM entity WHERE id = r.out)[0]
                LIMIT $limit"#,
            )
            .bind(("limit", limit))
            .await?
            .take(0)
            .unwrap_or_default();

        let mut node_map: std::collections::HashMap<String, GraphNode> =
            std::collections::HashMap::new();
        let mut edges = Vec::new();

        for row in rows {
            if row.a_uuid.is_empty() || row.b_uuid.is_empty() {
                continue;
            }
            node_map
                .entry(row.a_uuid.clone())
                .or_insert_with(|| GraphNode {
                    id: row.a_uuid.clone(),
                    name: row.a_name,
                    node_type: row.a_type,
                    summary: if row.a_summary.is_empty() {
                        None
                    } else {
                        Some(row.a_summary)
                    },
                });
            node_map
                .entry(row.b_uuid.clone())
                .or_insert_with(|| GraphNode {
                    id: row.b_uuid.clone(),
                    name: row.b_name,
                    node_type: row.b_type,
                    summary: if row.b_summary.is_empty() {
                        None
                    } else {
                        Some(row.b_summary)
                    },
                });
            edges.push(GraphEdge {
                id: row.r_uuid,
                source: row.a_uuid,
                target: row.b_uuid,
                edge_type: row.r_type,
                fact: if row.r_fact.is_empty() {
                    None
                } else {
                    Some(row.r_fact)
                },
                weight: 1.0,
            });
        }

        Ok(GraphData {
            nodes: node_map.into_values().collect(),
            edges,
        })
    }

    pub async fn get_node(&self, id: &str, _depth: i32) -> Result<NodeDetail> {
        let id_owned = id.to_string();

        let nodes: Vec<NodeRecord> = self
            .db
            .query("SELECT id, uuid, name, type, summary FROM entity WHERE uuid = $uuid LIMIT 1")
            .bind(("uuid", id_owned.clone()))
            .await?
            .take(0)
            .unwrap_or_default();

        let node_data = nodes
            .first()
            .ok_or_else(|| anyhow::anyhow!("node not found: {id}"))?;

        let node = GraphNode {
            id: node_data.uuid.clone(),
            name: node_data.name.clone(),
            node_type: node_data.node_type.clone(),
            summary: if node_data.summary.is_empty() {
                None
            } else {
                Some(node_data.summary.clone())
            },
        };

        let record_id = node_data.id.to_string();
        let rel_rows: Vec<RelRow> = self
            .db
            .query(
                r#"SELECT
                    r.uuid AS r_uuid, r.type AS r_type, r.fact AS r_fact,
                    n.uuid AS n_uuid, n.name AS n_name, n.type AS n_type, n.summary AS n_summary,
                    r.in = type::thing($record_id) AS is_outgoing
                FROM relates_to AS r
                WHERE r.in = type::thing($record_id) OR r.out = type::thing($record_id)
                LET n = IF r.in = type::thing($record_id) THEN (SELECT * FROM entity WHERE id = r.out)[0] ELSE (SELECT * FROM entity WHERE id = r.in)[0] END"#,
            )
            .bind(("record_id", record_id))
            .await?
            .take(0)
            .unwrap_or_default();

        let mut neighbors = Vec::new();
        let mut edges = Vec::new();
        let mut seen = std::collections::HashSet::new();

        for row in rel_rows {
            if row.n_uuid.is_empty() {
                continue;
            }
            if seen.insert(row.n_uuid.clone()) {
                neighbors.push(GraphNode {
                    id: row.n_uuid.clone(),
                    name: row.n_name,
                    node_type: row.n_type,
                    summary: if row.n_summary.is_empty() {
                        None
                    } else {
                        Some(row.n_summary)
                    },
                });
            }

            let (src, tgt) = if row.is_outgoing {
                (id_owned.clone(), row.n_uuid.clone())
            } else {
                (row.n_uuid.clone(), id_owned.clone())
            };

            edges.push(GraphEdge {
                id: row.r_uuid,
                source: src,
                target: tgt,
                edge_type: row.r_type,
                fact: if row.r_fact.is_empty() {
                    None
                } else {
                    Some(row.r_fact)
                },
                weight: 1.0,
            });
        }

        Ok(NodeDetail {
            node,
            neighbors,
            edges,
        })
    }

    pub fn db(&self) -> &Surreal<surrealdb::engine::local::Db> {
        &self.db
    }
}

// ── Internal SurrealDB row types ────────────────────────────────────

#[derive(Debug, serde::Deserialize)]
struct EntityRecord {
    id: surrealdb::sql::Thing,
    uuid: String,
}

#[derive(Debug, serde::Deserialize)]
struct EpisodeRecord {
    id: surrealdb::sql::Thing,
}

#[allow(dead_code)]
#[derive(Debug, serde::Deserialize)]
struct FactRow {
    #[serde(default)]
    r_uuid: String,
    #[serde(default)]
    r_type: String,
    #[serde(default)]
    r_fact: String,
    #[serde(default)]
    src_uuid: String,
    #[serde(default)]
    src_name: String,
    #[serde(default)]
    src_type: String,
    #[serde(default)]
    src_summary: String,
    #[serde(default)]
    tgt_uuid: String,
    #[serde(default)]
    tgt_name: String,
    #[serde(default)]
    tgt_type: String,
    #[serde(default)]
    tgt_summary: String,
    #[serde(default)]
    score: f64,
}

#[derive(Debug, serde::Deserialize)]
struct GraphRow {
    #[serde(default)]
    a_uuid: String,
    #[serde(default)]
    a_name: String,
    #[serde(default)]
    a_type: String,
    #[serde(default)]
    a_summary: String,
    #[serde(default)]
    r_uuid: String,
    #[serde(default)]
    r_type: String,
    #[serde(default)]
    r_fact: String,
    #[serde(default)]
    b_uuid: String,
    #[serde(default)]
    b_name: String,
    #[serde(default)]
    b_type: String,
    #[serde(default)]
    b_summary: String,
}

#[derive(Debug, serde::Deserialize)]
struct NodeRecord {
    id: surrealdb::sql::Thing,
    uuid: String,
    name: String,
    #[serde(rename = "type")]
    node_type: String,
    #[serde(default)]
    summary: String,
}

#[derive(Debug, serde::Deserialize)]
struct RelRow {
    #[serde(default)]
    r_uuid: String,
    #[serde(default)]
    r_type: String,
    #[serde(default)]
    r_fact: String,
    #[serde(default)]
    n_uuid: String,
    #[serde(default)]
    n_name: String,
    #[serde(default)]
    n_type: String,
    #[serde(default)]
    n_summary: String,
    #[serde(default)]
    is_outgoing: bool,
}
