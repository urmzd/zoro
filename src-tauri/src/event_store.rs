use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use surrealdb::engine::local::Db;
use surrealdb::Surreal;

use crate::models::*;

pub struct EventStore {
    db: Surreal<Db>,
}

impl EventStore {
    pub fn new(db: Surreal<Db>) -> Self {
        Self { db }
    }

    pub async fn ensure_schema(&self) -> Result<()> {
        let statements = [
            "DEFINE TABLE IF NOT EXISTS chat_session SCHEMAFULL",
            "DEFINE FIELD IF NOT EXISTS created_at ON chat_session TYPE datetime DEFAULT time::now()",
            "DEFINE TABLE IF NOT EXISTS chat_event SCHEMAFULL",
            "DEFINE FIELD IF NOT EXISTS session_id ON chat_event TYPE string",
            "DEFINE FIELD IF NOT EXISTS type ON chat_event TYPE string",
            "DEFINE FIELD IF NOT EXISTS role ON chat_event TYPE string",
            "DEFINE FIELD IF NOT EXISTS content ON chat_event TYPE string",
            "DEFINE FIELD IF NOT EXISTS tool_calls ON chat_event TYPE option<array>",
            "DEFINE FIELD IF NOT EXISTS created_at ON chat_event TYPE datetime DEFAULT time::now()",
            "DEFINE INDEX IF NOT EXISTS event_session ON chat_event FIELDS session_id",
            "DEFINE TABLE IF NOT EXISTS produced SCHEMAFULL TYPE RELATION IN chat_event OUT entity",
        ];

        for stmt in &statements {
            if let Err(e) = self.db.query(*stmt).await {
                log::warn!("event schema warning: {e} (statement: {stmt})");
            }
        }
        Ok(())
    }

    pub async fn create_session(&self) -> Result<String> {
        let session_id = uuid::Uuid::new_v4().to_string();

        self.db
            .query("CREATE type::thing('chat_session', $id) SET created_at = time::now()")
            .bind(("id", session_id.clone()))
            .await
            .context("create session")?;

        Ok(session_id)
    }

    pub async fn append_event(&self, session_id: &str, event: ChatEvent) -> Result<()> {
        let event_id = if event.id.is_empty() {
            uuid::Uuid::new_v4().to_string()
        } else {
            event.id
        };

        let has_tool_calls = !event.tool_calls.is_empty();

        if has_tool_calls {
            self.db
                .query("CREATE type::thing('chat_event', $id) SET session_id = $session_id, type = $type, role = $role, content = $content, tool_calls = $tool_calls")
                .bind(("id", event_id))
                .bind(("session_id", session_id.to_string()))
                .bind(("type", event.event_type))
                .bind(("role", event.role))
                .bind(("content", event.content))
                .bind(("tool_calls", event.tool_calls))
                .await
                .context("append event")?;
        } else {
            self.db
                .query("CREATE type::thing('chat_event', $id) SET session_id = $session_id, type = $type, role = $role, content = $content")
                .bind(("id", event_id))
                .bind(("session_id", session_id.to_string()))
                .bind(("type", event.event_type))
                .bind(("role", event.role))
                .bind(("content", event.content))
                .await
                .context("append event")?;
        }

        Ok(())
    }

    pub async fn get_session(&self, session_id: &str) -> Result<ChatSession> {
        let sess_rows: Vec<SessionRow> = self
            .db
            .query("SELECT created_at FROM type::thing('chat_session', $id)")
            .bind(("id", session_id.to_string()))
            .await?
            .take(0)
            .unwrap_or_default();

        let sess = sess_rows
            .first()
            .ok_or_else(|| anyhow::anyhow!("session not found: {session_id}"))?;

        let event_rows: Vec<EventRow> = self
            .db
            .query("SELECT type, role, content, tool_calls, created_at FROM chat_event WHERE session_id = $id ORDER BY created_at ASC")
            .bind(("id", session_id.to_string()))
            .await?
            .take(0)
            .unwrap_or_default();

        let messages: Vec<ChatMessage> = event_rows
            .into_iter()
            .map(|evt| ChatMessage {
                role: evt.role,
                content: evt.content,
                tool_calls: evt.tool_calls.unwrap_or_default(),
            })
            .collect();

        Ok(ChatSession {
            id: session_id.to_string(),
            messages,
            created_at: sess.created_at,
        })
    }

    pub async fn list_sessions(&self) -> Result<Vec<ChatSessionSummary>> {
        let sess_rows: Vec<SessionListRow> = self
            .db
            .query("SELECT id, created_at FROM chat_session ORDER BY created_at DESC")
            .await?
            .take(0)
            .unwrap_or_default();

        let mut summaries = Vec::new();
        for sess in sess_rows {
            let sess_id = extract_record_id(&sess.id.to_string());

            let preview_rows: Vec<PreviewRow> = self
                .db
                .query(
                    r#"SELECT content,
                        (SELECT VALUE count() FROM chat_event WHERE session_id = $id GROUP ALL) AS total
                    FROM chat_event
                    WHERE session_id = $id AND role = 'user'
                    ORDER BY created_at ASC LIMIT 1"#,
                )
                .bind(("id", sess_id.clone()))
                .await?
                .take(0)
                .unwrap_or_default();

            let (preview, msg_count) = if let Some(p) = preview_rows.first() {
                let mut preview = p.content.clone();
                if preview.len() > 120 {
                    preview.truncate(120);
                    preview.push_str("...");
                }
                (preview, p.total.unwrap_or(0))
            } else {
                (String::new(), 0)
            };

            summaries.push(ChatSessionSummary {
                id: sess_id,
                preview,
                message_count: msg_count,
                created_at: sess.created_at,
            });
        }

        Ok(summaries)
    }
}

fn extract_record_id(record_id: &str) -> String {
    if let Some(pos) = record_id.rfind(':') {
        record_id[pos + 1..].to_string()
    } else {
        record_id.to_string()
    }
}

#[derive(Debug, serde::Deserialize)]
struct SessionRow {
    created_at: DateTime<Utc>,
}

#[derive(Debug, serde::Deserialize)]
struct SessionListRow {
    id: surrealdb::sql::Thing,
    created_at: DateTime<Utc>,
}

#[allow(dead_code)]
#[derive(Debug, serde::Deserialize)]
struct EventRow {
    #[serde(default)]
    role: String,
    #[serde(default)]
    content: String,
    #[serde(default)]
    tool_calls: Option<Vec<ToolCall>>,
    #[serde(default)]
    created_at: DateTime<Utc>,
}

#[derive(Debug, serde::Deserialize)]
struct PreviewRow {
    #[serde(default)]
    content: String,
    #[serde(default)]
    total: Option<i64>,
}
