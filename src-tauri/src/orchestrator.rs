use std::fmt::Write;
use std::sync::Arc;
use tokio::sync::{mpsc, RwLock};

use crate::knowledge::KnowledgeStore;
use crate::models::*;
use crate::ollama::OllamaClient;
use crate::searcher::Searcher;

type Sessions = Arc<RwLock<std::collections::HashMap<String, ResearchSession>>>;
type Subscribers = Arc<RwLock<std::collections::HashMap<String, Vec<mpsc::Sender<SSEEvent>>>>>;

pub struct Orchestrator {
    knowledge: Arc<KnowledgeStore>,
    ollama: Arc<OllamaClient>,
    searcher: Arc<Searcher>,
    sessions: Sessions,
    subscribers: Subscribers,
}

impl Orchestrator {
    pub fn new(
        knowledge: Arc<KnowledgeStore>,
        ollama: Arc<OllamaClient>,
        searcher: Arc<Searcher>,
    ) -> Self {
        Self {
            knowledge,
            ollama,
            searcher,
            sessions: Arc::new(RwLock::new(std::collections::HashMap::new())),
            subscribers: Arc::new(RwLock::new(std::collections::HashMap::new())),
        }
    }

    pub async fn create_session(&self, query: &str) -> ResearchSession {
        let session = ResearchSession {
            id: uuid::Uuid::new_v4().to_string(),
            query: query.to_string(),
            status: "created".into(),
            results: vec![],
            entities: vec![],
            relations: vec![],
            timeline: vec![],
            summary: String::new(),
            created_at: chrono::Utc::now(),
        };

        self.sessions
            .write()
            .await
            .insert(session.id.clone(), session.clone());
        session
    }

    pub async fn get_session(&self, id: &str) -> Option<ResearchSession> {
        self.sessions.read().await.get(id).cloned()
    }

    pub async fn subscribe(&self, id: &str) -> Option<mpsc::Receiver<SSEEvent>> {
        let sessions = self.sessions.read().await;
        if !sessions.contains_key(id) {
            return None;
        }
        drop(sessions);

        let (tx, rx) = mpsc::channel(128);
        self.subscribers
            .write()
            .await
            .entry(id.to_string())
            .or_default()
            .push(tx);
        Some(rx)
    }

    async fn emit(&self, session_id: &str, evt: SSEEvent) {
        // Update timeline
        {
            let mut sessions = self.sessions.write().await;
            if let Some(session) = sessions.get_mut(session_id) {
                session.timeline.push(TimelineEvent {
                    event_type: evt.event_type.clone(),
                    message: format!("{}", evt.data),
                    timestamp: chrono::Utc::now(),
                    data: None,
                });
            }
        }

        let subs = self.subscribers.read().await;
        if let Some(channels) = subs.get(session_id) {
            for ch in channels {
                let _ = ch.try_send(evt.clone());
            }
        }
    }

    async fn close_subscribers(&self, session_id: &str) {
        self.subscribers.write().await.remove(session_id);
    }

    async fn set_status(&self, session_id: &str, status: &str) {
        let mut sessions = self.sessions.write().await;
        if let Some(s) = sessions.get_mut(session_id) {
            s.status = status.to_string();
        }
    }

    pub async fn run(&self, session_id: &str) {
        self.set_status(session_id, "running").await;

        let query = match self.get_session(session_id).await {
            Some(s) => s.query,
            None => {
                self.close_subscribers(session_id).await;
                return;
            }
        };

        // Step 1: Query knowledge store for prior knowledge
        self.emit(
            session_id,
            SSEEvent {
                event_type: EVENT_PRIOR_KNOWLEDGE.into(),
                data: serde_json::json!({"message": "Searching prior knowledge..."}),
            },
        )
        .await;

        match self.knowledge.search_facts(&query, "", &self.ollama).await {
            Ok(prior_facts) if !prior_facts.facts.is_empty() => {
                self.emit(
                    session_id,
                    SSEEvent {
                        event_type: EVENT_PRIOR_KNOWLEDGE.into(),
                        data: serde_json::to_value(&prior_facts.facts).unwrap_or_default(),
                    },
                )
                .await;
            }
            Err(e) => log::warn!("prior knowledge search error (non-fatal): {e}"),
            _ => {}
        }

        // Step 2: Web search
        self.emit(
            session_id,
            SSEEvent {
                event_type: EVENT_SEARCH_STARTED.into(),
                data: serde_json::json!({"query": query}),
            },
        )
        .await;

        let results = match self.searcher.search(&query).await {
            Ok(r) => r,
            Err(e) => {
                self.emit(
                    session_id,
                    SSEEvent {
                        event_type: EVENT_ERROR.into(),
                        data: serde_json::json!({"error": e.to_string()}),
                    },
                )
                .await;
                self.set_status(session_id, "error").await;
                self.close_subscribers(session_id).await;
                return;
            }
        };

        {
            let mut sessions = self.sessions.write().await;
            if let Some(s) = sessions.get_mut(session_id) {
                s.results = results.clone();
            }
        }

        self.emit(
            session_id,
            SSEEvent {
                event_type: EVENT_SEARCH_RESULTS.into(),
                data: serde_json::to_value(&results).unwrap_or_default(),
            },
        )
        .await;

        // Step 3: Ingest each result into knowledge store
        for (i, result) in results.iter().enumerate() {
            let episode_body = format!(
                "Title: {}\nURL: {}\nSnippet: {}",
                result.title, result.url, result.snippet
            );
            let req = EpisodeRequest {
                name: format!("{} - Result {}", query, i + 1),
                body: episode_body,
                source: result.url.clone(),
                group_id: session_id.to_string(),
            };

            match self.knowledge.add_episode(&req, &self.ollama).await {
                Ok(resp) => {
                    self.emit(
                        session_id,
                        SSEEvent {
                            event_type: EVENT_EPISODE_INGESTED.into(),
                            data: serde_json::json!({
                                "result_index": i,
                                "episode_uuid": resp.uuid,
                            }),
                        },
                    )
                    .await;

                    let mut sessions = self.sessions.write().await;
                    if let Some(s) = sessions.get_mut(session_id) {
                        for entity in &resp.entity_nodes {
                            s.entities.push(entity.clone());
                        }
                        for relation in &resp.episodic_edges {
                            s.relations.push(relation.clone());
                        }
                    }
                    drop(sessions);

                    for entity in &resp.entity_nodes {
                        self.emit(
                            session_id,
                            SSEEvent {
                                event_type: EVENT_ENTITY_DISCOVERED.into(),
                                data: serde_json::to_value(entity).unwrap_or_default(),
                            },
                        )
                        .await;
                    }
                    for relation in &resp.episodic_edges {
                        self.emit(
                            session_id,
                            SSEEvent {
                                event_type: EVENT_RELATION_FOUND.into(),
                                data: serde_json::to_value(relation).unwrap_or_default(),
                            },
                        )
                        .await;
                    }
                }
                Err(e) => {
                    log::warn!("episode ingestion error (result {}): {e}", i + 1);
                }
            }
        }

        // Step 4: Get session subgraph
        match self
            .knowledge
            .search_facts(&query, session_id, &self.ollama)
            .await
        {
            Ok(subgraph) => {
                self.emit(
                    session_id,
                    SSEEvent {
                        event_type: EVENT_GRAPH_READY.into(),
                        data: serde_json::to_value(&subgraph).unwrap_or_default(),
                    },
                )
                .await;

                // Step 5: Stream LLM synthesis
                let prompt = build_summary_prompt(&query, &results, &subgraph);
                match self.ollama.generate_stream(&prompt) {
                    Ok(mut rx) => {
                        let mut summary = String::new();
                        while let Some(token) = rx.recv().await {
                            summary.push_str(&token);
                            self.emit(
                                session_id,
                                SSEEvent {
                                    event_type: EVENT_SUMMARY_TOKEN.into(),
                                    data: serde_json::json!({"token": token}),
                                },
                            )
                            .await;
                        }
                        let mut sessions = self.sessions.write().await;
                        if let Some(s) = sessions.get_mut(session_id) {
                            s.summary = summary;
                        }
                    }
                    Err(e) => log::warn!("llm stream error: {e}"),
                }
            }
            Err(e) => log::warn!("subgraph search error: {e}"),
        }

        // Step 6: Complete
        self.set_status(session_id, "complete").await;
        self.emit(
            session_id,
            SSEEvent {
                event_type: EVENT_RESEARCH_COMPLETE.into(),
                data: serde_json::json!({"session_id": session_id}),
            },
        )
        .await;
        self.close_subscribers(session_id).await;
    }
}

fn build_summary_prompt(
    query: &str,
    results: &[SearchResult],
    facts: &SearchFactsResponse,
) -> String {
    let mut b = String::new();
    b.push_str("You are a research assistant. Synthesize the following search results and knowledge graph facts into a comprehensive research summary.\n\n");
    let _ = write!(b, "Research Query: {query}\n\n");

    b.push_str("## Search Results\n");
    for (i, r) in results.iter().enumerate() {
        let _ = write!(
            b,
            "{}. **{}** ({})\n   {}\n\n",
            i + 1,
            r.title,
            r.url,
            r.snippet
        );
    }

    if !facts.facts.is_empty() {
        b.push_str("## Knowledge Graph Facts\n");
        for f in &facts.facts {
            let _ = writeln!(b, "- {}: {}", f.name, f.fact);
        }
        b.push('\n');
    }

    b.push_str("Provide a well-structured markdown summary with sections, key findings, and connections between topics. Include code examples where relevant.");
    b
}
