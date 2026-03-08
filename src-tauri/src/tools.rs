use anyhow::Result;
use std::collections::HashMap;
use std::fmt::Write;
use std::sync::Arc;

use crate::knowledge::KnowledgeStore;
use crate::models::*;
use crate::ollama::OllamaClient;
use crate::searcher::Searcher;

pub struct ToolRegistry {
    searcher: Arc<Searcher>,
    knowledge: Arc<KnowledgeStore>,
    ollama: Arc<OllamaClient>,
    store_group_id: String,
}

impl ToolRegistry {
    pub fn new(
        searcher: Arc<Searcher>,
        knowledge: Arc<KnowledgeStore>,
        ollama: Arc<OllamaClient>,
    ) -> Self {
        Self {
            searcher,
            knowledge,
            ollama,
            store_group_id: String::new(),
        }
    }

    pub fn clone_with_group(&self, group_id: &str) -> Self {
        Self {
            searcher: self.searcher.clone(),
            knowledge: self.knowledge.clone(),
            ollama: self.ollama.clone(),
            store_group_id: group_id.to_string(),
        }
    }

    pub fn definitions(&self) -> Vec<OllamaTool> {
        vec![
            OllamaTool {
                tool_type: "function".into(),
                function: OllamaToolFunction {
                    name: "web_search".into(),
                    description: "Search the web for current information on a topic. Returns up to 5 results with titles, URLs, and snippets.".into(),
                    parameters: OllamaToolFunctionParams {
                        param_type: "object".into(),
                        required: vec!["query".into()],
                        properties: HashMap::from([
                            ("query".into(), OllamaToolProperty {
                                prop_type: "string".into(),
                                description: "The search query".into(),
                            }),
                        ]),
                    },
                },
            },
            OllamaTool {
                tool_type: "function".into(),
                function: OllamaToolFunction {
                    name: "search_knowledge".into(),
                    description: "Search the knowledge graph for previously stored facts and entities. Returns up to 10 relevant facts.".into(),
                    parameters: OllamaToolFunctionParams {
                        param_type: "object".into(),
                        required: vec!["query".into()],
                        properties: HashMap::from([
                            ("query".into(), OllamaToolProperty {
                                prop_type: "string".into(),
                                description: "The search query for knowledge retrieval".into(),
                            }),
                        ]),
                    },
                },
            },
            OllamaTool {
                tool_type: "function".into(),
                function: OllamaToolFunction {
                    name: "store_knowledge".into(),
                    description: "Store information into the knowledge graph by extracting entities and relationships from text. Use this to persist important findings.".into(),
                    parameters: OllamaToolFunctionParams {
                        param_type: "object".into(),
                        required: vec!["text".into(), "source".into()],
                        properties: HashMap::from([
                            ("text".into(), OllamaToolProperty {
                                prop_type: "string".into(),
                                description: "The text content to extract knowledge from".into(),
                            }),
                            ("source".into(), OllamaToolProperty {
                                prop_type: "string".into(),
                                description: "Description of the source of this information".into(),
                            }),
                        ]),
                    },
                },
            },
        ]
    }

    pub async fn execute(
        &self,
        name: &str,
        args: &serde_json::Map<String, serde_json::Value>,
    ) -> Result<String> {
        match name {
            "web_search" => self.web_search(args).await,
            "search_knowledge" => self.search_knowledge(args).await,
            "store_knowledge" => self.store_knowledge(args).await,
            _ => anyhow::bail!("unknown tool: {name}"),
        }
    }

    async fn web_search(
        &self,
        args: &serde_json::Map<String, serde_json::Value>,
    ) -> Result<String> {
        let query = args
            .get("query")
            .and_then(|v| v.as_str())
            .unwrap_or("");
        if query.is_empty() {
            anyhow::bail!("web_search: query is required");
        }

        let results = self.searcher.search(query).await?;

        let mut out = String::new();
        let limit = results.len().min(5);
        for (i, r) in results.iter().take(limit).enumerate() {
            let snippet = if r.snippet.len() > 200 {
                format!("{}...", &r.snippet[..200])
            } else {
                r.snippet.clone()
            };
            writeln!(out, "{}. {}\n   {}\n   {}\n", i + 1, r.title, r.url, snippet)?;
        }

        if out.is_empty() {
            Ok("No results found.".into())
        } else {
            Ok(out)
        }
    }

    async fn search_knowledge(
        &self,
        args: &serde_json::Map<String, serde_json::Value>,
    ) -> Result<String> {
        let query = args
            .get("query")
            .and_then(|v| v.as_str())
            .unwrap_or("");
        if query.is_empty() {
            anyhow::bail!("search_knowledge: query is required");
        }

        let resp = self.knowledge.search_facts(query, "", &self.ollama).await?;

        let mut out = String::new();
        let limit = resp.facts.len().min(10);
        for f in resp.facts.iter().take(limit) {
            writeln!(
                out,
                "- {} -> {}: {}",
                f.source_node.name, f.target_node.name, f.fact
            )?;
        }

        if out.is_empty() {
            Ok("No relevant knowledge found.".into())
        } else {
            Ok(out)
        }
    }

    async fn store_knowledge(
        &self,
        args: &serde_json::Map<String, serde_json::Value>,
    ) -> Result<String> {
        let text = args
            .get("text")
            .and_then(|v| v.as_str())
            .unwrap_or("");
        if text.is_empty() {
            anyhow::bail!("store_knowledge: text is required");
        }
        let source = args
            .get("source")
            .and_then(|v| v.as_str())
            .unwrap_or("chat");

        let req = EpisodeRequest {
            name: source.to_string(),
            body: text.to_string(),
            source: source.to_string(),
            group_id: self.store_group_id.clone(),
        };

        let resp = self.knowledge.add_episode(&req, &self.ollama).await?;

        Ok(format!(
            "Stored {} entities and {} relations.",
            resp.entity_nodes.len(),
            resp.episodic_edges.len()
        ))
    }
}
