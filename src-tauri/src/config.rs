use serde::{Deserialize, Serialize};
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    pub ollama_host: String,
    pub ollama_model: String,
    pub ollama_fast_model: String,
    pub embedding_model: String,
    pub db_path: PathBuf,
}

impl Default for AppConfig {
    fn default() -> Self {
        let db_path = dirs::data_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("zoro")
            .join("zoro.db");

        Self {
            ollama_host: std::env::var("OLLAMA_HOST")
                .unwrap_or_else(|_| "http://localhost:11434".into()),
            ollama_model: std::env::var("OLLAMA_MODEL")
                .unwrap_or_else(|_| "qwen3.5:4b".into()),
            ollama_fast_model: std::env::var("OLLAMA_FAST_MODEL")
                .unwrap_or_else(|_| "qwen3.5:0.8b".into()),
            embedding_model: std::env::var("EMBEDDING_MODEL")
                .unwrap_or_else(|_| "nomic-embed-text".into()),
            db_path,
        }
    }
}
