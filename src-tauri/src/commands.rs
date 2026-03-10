use std::sync::Arc;
use tauri::{AppHandle, Emitter, State};

use crate::agent::Agent;
use crate::knowledge::KnowledgeStore;
use crate::models::*;
use crate::ollama::OllamaClient;
use crate::orchestrator::Orchestrator;

pub struct AppState {
    pub agent: Arc<Agent>,
    pub orchestrator: Arc<Orchestrator>,
    pub knowledge: Arc<KnowledgeStore>,
    pub ollama: Arc<OllamaClient>,
}

// ── Chat commands ───────────────────────────────────────────────────

#[tauri::command]
pub async fn create_chat_session(
    state: State<'_, AppState>,
) -> Result<ChatSession, String> {
    log::info!("[cmd] create_chat_session");
    let result = state.agent.create_session().await.map_err(|e| e.to_string());
    log::info!("[cmd] create_chat_session -> {:?}", result.is_ok());
    result
}

#[tauri::command]
pub async fn list_chat_sessions(
    state: State<'_, AppState>,
) -> Result<Vec<ChatSessionSummary>, String> {
    log::info!("[cmd] list_chat_sessions");
    let result = state.agent.list_sessions().await.map_err(|e| e.to_string());
    log::info!("[cmd] list_chat_sessions -> {:?}", result.is_ok());
    result
}

#[tauri::command]
pub async fn get_chat_session(
    state: State<'_, AppState>,
    id: String,
) -> Result<ChatSession, String> {
    log::info!("[cmd] get_chat_session id={id}");
    let result = state.agent.get_session(&id).await.map_err(|e| e.to_string());
    log::info!("[cmd] get_chat_session -> {:?}", result.is_ok());
    result
}

#[tauri::command]
pub async fn send_chat_message(
    app: AppHandle,
    state: State<'_, AppState>,
    id: String,
    content: String,
) -> Result<(), String> {
    log::info!("[cmd] send_chat_message id={id} content={:?}", &content[..content.len().min(100)]);
    let agent = state.agent.clone();
    let event_name = format!("chat-event:{id}");

    // Subscribe before sending to avoid race
    let mut rx = agent.subscribe(&id).await;

    // Run agent in background
    let agent_clone = agent.clone();
    let id_clone = id.clone();
    tokio::spawn(async move {
        agent_clone.send_message(&id_clone, &content).await;
    });

    // Forward events to frontend via Tauri events
    let app_clone = app.clone();
    tokio::spawn(async move {
        while let Some(evt) = rx.recv().await {
            let _ = app_clone.emit(&event_name, &evt);
        }
    });

    Ok(())
}

// ── Research commands ───────────────────────────────────────────────

#[tauri::command]
pub async fn start_research(
    app: AppHandle,
    state: State<'_, AppState>,
    query: String,
) -> Result<String, String> {
    log::info!("[cmd] start_research query={query:?}");
    let session = state.orchestrator.create_session(&query).await;
    let session_id = session.id.clone();
    let event_name = format!("research-event:{session_id}");

    // Subscribe
    let mut rx = state
        .orchestrator
        .subscribe(&session_id)
        .await
        .ok_or("session not found")?;

    // Run orchestrator in background
    let orchestrator = state.orchestrator.clone();
    let sid = session_id.clone();
    tokio::spawn(async move {
        orchestrator.run(&sid).await;
    });

    // Forward events
    let app_clone = app.clone();
    tokio::spawn(async move {
        while let Some(evt) = rx.recv().await {
            let _ = app_clone.emit(&event_name, &evt);
        }
    });

    Ok(session_id)
}

// ── Knowledge commands ──────────────────────────────────────────────

#[tauri::command]
pub async fn search_knowledge(
    state: State<'_, AppState>,
    query: String,
) -> Result<SearchFactsResponse, String> {
    log::info!("[cmd] search_knowledge query={query:?}");
    let result = state
        .knowledge
        .search_facts(&query, "", &state.ollama)
        .await
        .map_err(|e| e.to_string());
    log::info!("[cmd] search_knowledge -> {:?}", result.is_ok());
    result
}

#[tauri::command]
pub async fn get_knowledge_graph(
    state: State<'_, AppState>,
    limit: Option<i64>,
) -> Result<GraphData, String> {
    log::info!("[cmd] get_knowledge_graph limit={limit:?}");
    let result = state
        .knowledge
        .get_graph(limit.unwrap_or(300))
        .await
        .map_err(|e| e.to_string());
    log::info!("[cmd] get_knowledge_graph -> {:?}", result.is_ok());
    result
}

#[tauri::command]
pub async fn get_node_detail(
    state: State<'_, AppState>,
    id: String,
    depth: Option<i32>,
) -> Result<NodeDetail, String> {
    log::info!("[cmd] get_node_detail id={id}");
    let result = state
        .knowledge
        .get_node(&id, depth.unwrap_or(1))
        .await
        .map_err(|e| e.to_string());
    log::info!("[cmd] get_node_detail -> {:?}", result.is_ok());
    result
}

// ── Intent & Autocomplete commands ──────────────────────────────────

#[derive(serde::Serialize)]
pub struct IntentResponse {
    pub action: String,
    pub query: String,
}

#[tauri::command]
pub async fn classify_intent(
    state: State<'_, AppState>,
    query: String,
) -> Result<IntentResponse, String> {
    log::info!("[cmd] classify_intent query={query:?}");
    let action = state
        .agent
        .classify_intent(&query)
        .await
        .unwrap_or_else(|e| {
            log::warn!("[cmd] classify_intent error: {e}");
            "chat".into()
        });
    log::info!("[cmd] classify_intent -> action={action}");
    Ok(IntentResponse {
        action,
        query,
    })
}

#[derive(serde::Serialize)]
pub struct AutocompleteResponse {
    pub suggestions: Vec<String>,
}

#[tauri::command]
pub async fn get_autocomplete(
    state: State<'_, AppState>,
    query: String,
) -> Result<AutocompleteResponse, String> {
    log::info!("[cmd] get_autocomplete query={query:?}");
    let suggestions = state.agent.autocomplete(&query).await;
    log::info!("[cmd] get_autocomplete -> {} suggestions", suggestions.len());
    Ok(AutocompleteResponse { suggestions })
}
