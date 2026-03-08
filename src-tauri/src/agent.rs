use anyhow::Result;
use std::sync::Arc;
use tokio::sync::{mpsc, RwLock};

use crate::event_store::EventStore;
use crate::models::*;
use crate::ollama::OllamaClient;
use crate::tools::ToolRegistry;

const MAX_AGENT_ITERATIONS: usize = 10;

const SYSTEM_PROMPT: &str = r#"You are Zoro, an AI research assistant with access to tools. Your purpose is to help users understand topics by searching the web, querying a knowledge graph, and storing important findings.

When answering:
- Use web_search to find current information
- Use search_knowledge to check what's already known
- Use store_knowledge to persist important findings for future reference
- Synthesize information from multiple sources
- Be concise and well-structured in your responses
- Use markdown formatting"#;

type Subscribers = Arc<RwLock<std::collections::HashMap<String, Vec<mpsc::Sender<SSEEvent>>>>>;

pub struct Agent {
    ollama: Arc<OllamaClient>,
    tools: Arc<ToolRegistry>,
    registry: Arc<ModelRegistry>,
    events: Arc<EventStore>,
    subscribers: Subscribers,
}

impl Agent {
    pub fn new(
        ollama: Arc<OllamaClient>,
        tools: Arc<ToolRegistry>,
        registry: Arc<ModelRegistry>,
        events: Arc<EventStore>,
    ) -> Self {
        Self {
            ollama,
            tools,
            registry,
            events,
            subscribers: Arc::new(RwLock::new(std::collections::HashMap::new())),
        }
    }

    pub async fn create_session(&self) -> Result<ChatSession> {
        let id = self.events.create_session().await?;
        Ok(ChatSession {
            id,
            messages: vec![],
            created_at: chrono::Utc::now(),
        })
    }

    pub async fn get_session(&self, id: &str) -> Result<ChatSession> {
        self.events.get_session(id).await
    }

    pub async fn list_sessions(&self) -> Result<Vec<ChatSessionSummary>> {
        self.events.list_sessions().await
    }

    pub async fn subscribe(&self, id: &str) -> mpsc::Receiver<SSEEvent> {
        let (tx, rx) = mpsc::channel(128);
        let mut subs = self.subscribers.write().await;
        subs.entry(id.to_string()).or_default().push(tx);
        rx
    }

    async fn emit(&self, session_id: &str, evt: SSEEvent) {
        let subs = self.subscribers.read().await;
        if let Some(channels) = subs.get(session_id) {
            for ch in channels {
                let _ = ch.try_send(evt.clone());
            }
        }
    }

    async fn close_subscribers(&self, session_id: &str) {
        let mut subs = self.subscribers.write().await;
        subs.remove(session_id);
        // Senders are dropped, receivers will see channel closed
    }

    pub async fn send_message(&self, session_id: &str, content: &str) {
        // Append user message event
        let user_event = ChatEvent {
            id: String::new(),
            session_id: session_id.to_string(),
            event_type: "user_message".into(),
            role: "user".into(),
            content: content.to_string(),
            tool_calls: vec![],
            created_at: chrono::Utc::now(),
        };
        if let Err(e) = self.events.append_event(session_id, user_event).await {
            self.emit(
                session_id,
                SSEEvent {
                    event_type: EVENT_ERROR.into(),
                    data: serde_json::json!({"message": e.to_string()}),
                },
            )
            .await;
            self.close_subscribers(session_id).await;
            return;
        }

        // Clone tools with session binding
        let tools = self.tools.clone_with_group(session_id);

        self.run_loop(session_id, &tools).await;
        self.close_subscribers(session_id).await;
    }

    async fn run_loop(&self, session_id: &str, tools: &ToolRegistry) {
        for _iteration in 0..MAX_AGENT_ITERATIONS {
            // Rebuild messages from event store
            let session = match self.events.get_session(session_id).await {
                Ok(s) => s,
                Err(e) => {
                    self.emit(
                        session_id,
                        SSEEvent {
                            event_type: EVENT_ERROR.into(),
                            data: serde_json::json!({"message": e.to_string()}),
                        },
                    )
                    .await;
                    return;
                }
            };

            let ollama_messages = build_ollama_messages(&session.messages);
            let tool_defs = tools.definitions();

            let mut rx = match self.ollama.chat_stream(ollama_messages, tool_defs) {
                Ok(rx) => rx,
                Err(e) => {
                    log::error!("agent llm error: {e}");
                    self.emit(
                        session_id,
                        SSEEvent {
                            event_type: EVENT_ERROR.into(),
                            data: serde_json::json!({"message": e.to_string()}),
                        },
                    )
                    .await;
                    return;
                }
            };

            // Accumulate the full assistant response
            let mut content_builder = String::new();
            let mut tool_calls: Vec<OllamaToolCall> = Vec::new();

            while let Some(chunk) = rx.recv().await {
                if !chunk.message.content.is_empty() {
                    content_builder.push_str(&chunk.message.content);
                    self.emit(
                        session_id,
                        SSEEvent {
                            event_type: EVENT_TEXT_DELTA.into(),
                            data: serde_json::json!({"content": chunk.message.content}),
                        },
                    )
                    .await;
                }
                if !chunk.message.tool_calls.is_empty() {
                    tool_calls.extend(chunk.message.tool_calls);
                }
            }

            // Build tool calls for the event
            let event_tool_calls: Vec<ToolCall> = tool_calls
                .iter()
                .map(|tc| ToolCall {
                    id: uuid::Uuid::new_v4().to_string(),
                    name: tc.function.name.clone(),
                    arguments: serde_json::to_string(&tc.function.arguments).unwrap_or_default(),
                    result: None,
                })
                .collect();

            // Append assistant message event
            let assistant_event = ChatEvent {
                id: String::new(),
                session_id: session_id.to_string(),
                event_type: "assistant_message".into(),
                role: "assistant".into(),
                content: content_builder,
                tool_calls: event_tool_calls.clone(),
                created_at: chrono::Utc::now(),
            };
            if let Err(e) = self.events.append_event(session_id, assistant_event).await {
                log::error!("append assistant event error: {e}");
            }

            // If no tool calls, we're done
            if tool_calls.is_empty() {
                break;
            }

            // Execute tool calls and append results
            for (i, tc) in tool_calls.iter().enumerate() {
                let call_id = &event_tool_calls[i].id;
                let args_json = serde_json::to_string(&tc.function.arguments).unwrap_or_default();

                self.emit(
                    session_id,
                    SSEEvent {
                        event_type: EVENT_TOOL_CALL_START.into(),
                        data: serde_json::json!({
                            "id": call_id,
                            "name": tc.function.name,
                            "arguments": args_json,
                        }),
                    },
                )
                .await;

                let result = match tools.execute(&tc.function.name, &tc.function.arguments).await {
                    Ok(r) => r,
                    Err(e) => format!("Error: {e}"),
                };

                self.emit(
                    session_id,
                    SSEEvent {
                        event_type: EVENT_TOOL_CALL_RESULT.into(),
                        data: serde_json::json!({
                            "id": call_id,
                            "name": tc.function.name,
                            "result": result,
                        }),
                    },
                )
                .await;

                // Append tool result event
                let tool_event = ChatEvent {
                    id: String::new(),
                    session_id: session_id.to_string(),
                    event_type: "tool_message".into(),
                    role: "tool".into(),
                    content: result,
                    tool_calls: vec![],
                    created_at: chrono::Utc::now(),
                };
                if let Err(e) = self.events.append_event(session_id, tool_event).await {
                    log::error!("append tool event error: {e}");
                }
            }
        }

        self.emit(
            session_id,
            SSEEvent {
                event_type: EVENT_DONE.into(),
                data: serde_json::Value::Null,
            },
        )
        .await;
    }

    pub async fn classify_intent(&self, query: &str) -> Result<String> {
        let prompt = format!(
            r#"Classify the user's intent as "chat" or "knowledge_search".

"knowledge_search": user wants to look up or find something stored in the knowledge graph.
"chat": user wants to research something new, have a conversation, or ask a general question.

User query: {query}"#
        );

        let format = serde_json::json!({
            "type": "object",
            "properties": {
                "action": {
                    "type": "string",
                    "enum": ["chat", "knowledge_search"]
                }
            },
            "required": ["action"]
        });

        let resp = self
            .ollama
            .generate_with_model(
                &prompt,
                self.registry.model(ModelTier::Fast),
                Some(format),
                None,
            )
            .await?;

        #[derive(serde::Deserialize)]
        struct IntentResult {
            action: String,
        }

        match serde_json::from_str::<IntentResult>(&resp) {
            Ok(r) if r.action == "knowledge_search" => Ok("knowledge_search".into()),
            _ => Ok("chat".into()),
        }
    }

    pub async fn autocomplete(&self, query: &str) -> Vec<String> {
        let prompt = format!(
            r#"Given the partial query below, suggest 3 to 5 complete search queries the user might intend. Return ONLY a JSON array of strings, no extra text.

Partial query: {query}"#
        );

        let raw = match self
            .ollama
            .generate_with_model(&prompt, self.registry.model(ModelTier::Fast), None, None)
            .await
        {
            Ok(r) => r,
            Err(_) => return vec![],
        };

        parse_json_array(&raw)
    }
}

fn build_ollama_messages(msgs: &[ChatMessage]) -> Vec<OllamaChatMessage> {
    let mut ollama_msgs = vec![OllamaChatMessage {
        role: "system".into(),
        content: SYSTEM_PROMPT.into(),
        tool_calls: vec![],
    }];

    for m in msgs {
        let mut om = OllamaChatMessage {
            role: m.role.clone(),
            content: m.content.clone(),
            tool_calls: vec![],
        };
        if m.role == "assistant" && !m.tool_calls.is_empty() {
            for tc in &m.tool_calls {
                let args: serde_json::Map<String, serde_json::Value> =
                    serde_json::from_str(&tc.arguments).unwrap_or_default();
                om.tool_calls.push(OllamaToolCall {
                    function: OllamaToolCallFunction {
                        name: tc.name.clone(),
                        arguments: args,
                    },
                });
            }
        }
        ollama_msgs.push(om);
    }

    ollama_msgs
}

fn parse_json_array(raw: &str) -> Vec<String> {
    let start = match raw.find('[') {
        Some(s) => s,
        None => return vec![],
    };
    let end = match raw.rfind(']') {
        Some(e) => e,
        None => return vec![],
    };
    if end <= start {
        return vec![];
    }
    serde_json::from_str::<Vec<String>>(&raw[start..=end]).unwrap_or_default()
}
