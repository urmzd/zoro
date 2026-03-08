use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

// ── Chat types ──────────────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatSession {
    pub id: String,
    pub messages: Vec<ChatMessage>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatMessage {
    pub role: String,
    pub content: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub tool_calls: Vec<ToolCall>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    pub id: String,
    pub name: String,
    pub arguments: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub result: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatSessionSummary {
    pub id: String,
    pub preview: String,
    pub message_count: i64,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatEvent {
    #[serde(default)]
    pub id: String,
    #[serde(default)]
    pub session_id: String,
    #[serde(rename = "type")]
    pub event_type: String,
    pub role: String,
    pub content: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub tool_calls: Vec<ToolCall>,
    #[serde(default)]
    pub created_at: DateTime<Utc>,
}

// ── SSE event types ─────────────────────────────────────────────────

pub const EVENT_SEARCH_STARTED: &str = "search_started";
pub const EVENT_SEARCH_RESULTS: &str = "search_results";
pub const EVENT_EPISODE_INGESTED: &str = "episode_ingested";
pub const EVENT_ENTITY_DISCOVERED: &str = "entity_discovered";
pub const EVENT_RELATION_FOUND: &str = "relation_found";
pub const EVENT_PRIOR_KNOWLEDGE: &str = "prior_knowledge";
pub const EVENT_GRAPH_READY: &str = "graph_ready";
pub const EVENT_SUMMARY_TOKEN: &str = "summary_token";
pub const EVENT_RESEARCH_COMPLETE: &str = "research_complete";
pub const EVENT_ERROR: &str = "error";
pub const EVENT_TEXT_DELTA: &str = "text_delta";
pub const EVENT_TOOL_CALL_START: &str = "tool_call_start";
pub const EVENT_TOOL_CALL_RESULT: &str = "tool_call_result";
pub const EVENT_DONE: &str = "done";

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SSEEvent {
    #[serde(rename = "type")]
    pub event_type: String,
    pub data: serde_json::Value,
}

// ── Knowledge types ─────────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EpisodeRequest {
    pub name: String,
    #[serde(rename = "episode_body")]
    pub body: String,
    #[serde(rename = "source_description")]
    pub source: String,
    pub group_id: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EpisodeResponse {
    pub uuid: String,
    pub name: String,
    #[serde(default)]
    pub entity_nodes: Vec<Entity>,
    #[serde(default)]
    pub episodic_edges: Vec<Relation>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SearchFactsResponse {
    pub facts: Vec<Fact>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Fact {
    pub uuid: String,
    pub name: String,
    pub fact: String,
    pub source_node: Entity,
    pub target_node: Entity,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphData {
    pub nodes: Vec<GraphNode>,
    pub edges: Vec<GraphEdge>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphNode {
    pub id: String,
    pub name: String,
    #[serde(rename = "type")]
    pub node_type: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub summary: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphEdge {
    pub id: String,
    pub source: String,
    pub target: String,
    #[serde(rename = "type")]
    pub edge_type: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub fact: Option<String>,
    pub weight: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeDetail {
    pub node: GraphNode,
    pub neighbors: Vec<GraphNode>,
    pub edges: Vec<GraphEdge>,
}

// ── Research types ──────────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResearchSession {
    pub id: String,
    pub query: String,
    pub status: String,
    pub results: Vec<SearchResult>,
    pub entities: Vec<Entity>,
    pub relations: Vec<Relation>,
    pub timeline: Vec<TimelineEvent>,
    pub summary: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SearchResult {
    pub title: String,
    pub url: String,
    pub snippet: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Entity {
    pub uuid: String,
    pub name: String,
    #[serde(rename = "type")]
    pub entity_type: String,
    pub summary: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Relation {
    pub uuid: String,
    pub source_uuid: String,
    pub target_uuid: String,
    #[serde(rename = "type")]
    pub relation_type: String,
    pub fact: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimelineEvent {
    #[serde(rename = "type")]
    pub event_type: String,
    pub message: String,
    pub timestamp: DateTime<Utc>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub data: Option<serde_json::Value>,
}

// ── Ollama wire types ───────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaChatMessage {
    pub role: String,
    pub content: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub tool_calls: Vec<OllamaToolCall>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaToolCall {
    pub function: OllamaToolCallFunction,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaToolCallFunction {
    pub name: String,
    pub arguments: serde_json::Map<String, serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaTool {
    #[serde(rename = "type")]
    pub tool_type: String,
    pub function: OllamaToolFunction,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaToolFunction {
    pub name: String,
    pub description: String,
    pub parameters: OllamaToolFunctionParams,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaToolFunctionParams {
    #[serde(rename = "type")]
    pub param_type: String,
    pub required: Vec<String>,
    pub properties: std::collections::HashMap<String, OllamaToolProperty>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaToolProperty {
    #[serde(rename = "type")]
    pub prop_type: String,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaChatRequest {
    pub model: String,
    pub messages: Vec<OllamaChatMessage>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub tools: Vec<OllamaTool>,
    pub stream: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaChatChunk {
    pub message: OllamaChatMessage,
    #[serde(default)]
    pub done: bool,
}

// ── Ollama generate types ───────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaGenerateRequest {
    pub model: String,
    pub prompt: String,
    pub stream: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub format: Option<serde_json::Value>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub options: Option<serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaGenerateResponse {
    #[serde(default)]
    pub response: String,
    #[serde(default)]
    pub done: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaEmbedRequest {
    pub model: String,
    pub input: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaEmbedResponse {
    pub embeddings: Vec<Vec<f32>>,
}

// ── Extraction types ────────────────────────────────────────────────

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExtractedEntity {
    pub name: String,
    #[serde(rename = "type")]
    pub entity_type: String,
    pub summary: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExtractedRelation {
    pub source: String,
    pub target: String,
    #[serde(rename = "type")]
    pub relation_type: String,
    pub fact: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExtractedData {
    pub entities: Vec<ExtractedEntity>,
    pub relations: Vec<ExtractedRelation>,
}

// ── Model registry ──────────────────────────────────────────────────

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum ModelTier {
    Standard,
    Fast,
    Embedding,
}

#[derive(Debug, Clone)]
pub struct ModelRegistry {
    pub standard: String,
    pub fast: String,
    pub embedding: String,
}

impl ModelRegistry {
    pub fn new(standard: String, fast: String, embedding: String) -> Self {
        let fast = if fast.is_empty() {
            standard.clone()
        } else {
            fast
        };
        Self {
            standard,
            fast,
            embedding,
        }
    }

    pub fn model(&self, tier: ModelTier) -> &str {
        match tier {
            ModelTier::Standard => &self.standard,
            ModelTier::Fast => &self.fast,
            ModelTier::Embedding => &self.embedding,
        }
    }
}
