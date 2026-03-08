use anyhow::{bail, Context, Result};
use reqwest::Client;
use std::time::Duration;
use tokio::io::AsyncBufReadExt;
use tokio::sync::mpsc;

use crate::models::*;

pub struct OllamaClient {
    host: String,
    model: String,
    embedding_model: String,
    http: Client,
}

impl OllamaClient {
    pub fn new(host: &str, model: &str, embedding_model: &str) -> Self {
        Self {
            host: host.to_string(),
            model: model.to_string(),
            embedding_model: embedding_model.to_string(),
            http: Client::builder()
                .timeout(Duration::from_secs(300))
                .build()
                .expect("http client"),
        }
    }

    pub async fn generate(&self, prompt: &str) -> Result<String> {
        self.generate_with_model(prompt, &self.model, None, None)
            .await
    }

    pub async fn generate_with_model(
        &self,
        prompt: &str,
        model: &str,
        format: Option<serde_json::Value>,
        options: Option<serde_json::Value>,
    ) -> Result<String> {
        let req = OllamaGenerateRequest {
            model: model.to_string(),
            prompt: prompt.to_string(),
            stream: false,
            format,
            options,
        };

        let resp = self
            .http
            .post(format!("{}/api/generate", self.host))
            .json(&req)
            .send()
            .await
            .context("ollama generate")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("ollama returned {status}: {body}");
        }

        let result: OllamaGenerateResponse = resp.json().await.context("decode ollama response")?;
        Ok(result.response)
    }

    pub fn generate_stream(
        &self,
        prompt: &str,
    ) -> Result<mpsc::Receiver<String>> {
        let (tx, rx) = mpsc::channel(64);
        let url = format!("{}/api/generate", self.host);
        let req = OllamaGenerateRequest {
            model: self.model.clone(),
            prompt: prompt.to_string(),
            stream: true,
            format: None,
            options: None,
        };
        let http = self.http.clone();

        tokio::spawn(async move {
            let resp = match http.post(&url).json(&req).send().await {
                Ok(r) if r.status().is_success() => r,
                _ => return,
            };

            let stream = resp.bytes_stream();
            use tokio_stream::StreamExt;
            use tokio_util::io::StreamReader;

            let reader = StreamReader::new(
                stream.map(|r| r.map_err(std::io::Error::other)),
            );
            let mut lines = tokio::io::BufReader::new(reader).lines();

            while let Ok(Some(line)) = lines.next_line().await {
                if line.is_empty() {
                    continue;
                }
                if let Ok(chunk) = serde_json::from_str::<OllamaGenerateResponse>(&line) {
                    if !chunk.response.is_empty() && tx.send(chunk.response).await.is_err() {
                        return;
                    }
                    if chunk.done {
                        return;
                    }
                }
            }
        });

        Ok(rx)
    }

    pub async fn embed(&self, text: &str) -> Result<Vec<f32>> {
        let req = OllamaEmbedRequest {
            model: self.embedding_model.clone(),
            input: text.to_string(),
        };

        let resp = self
            .http
            .post(format!("{}/api/embed", self.host))
            .json(&req)
            .send()
            .await
            .context("ollama embed")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("ollama embed returned {status}: {body}");
        }

        let result: OllamaEmbedResponse = resp.json().await.context("decode embed response")?;
        result
            .embeddings
            .into_iter()
            .next()
            .ok_or_else(|| anyhow::anyhow!("no embeddings returned"))
    }

    pub fn chat_stream(
        &self,
        messages: Vec<OllamaChatMessage>,
        tools: Vec<OllamaTool>,
    ) -> Result<mpsc::Receiver<OllamaChatChunk>> {
        let (tx, rx) = mpsc::channel(64);
        let url = format!("{}/api/chat", self.host);
        let req = OllamaChatRequest {
            model: self.model.clone(),
            messages,
            tools,
            stream: true,
        };
        let http = self.http.clone();

        tokio::spawn(async move {
            let resp = match http.post(&url).json(&req).send().await {
                Ok(r) if r.status().is_success() => r,
                _ => return,
            };

            let stream = resp.bytes_stream();
            use tokio_stream::StreamExt;
            use tokio_util::io::StreamReader;

            let reader = StreamReader::new(
                stream.map(|r| r.map_err(std::io::Error::other)),
            );
            let mut lines = tokio::io::BufReader::new(reader).lines();

            while let Ok(Some(line)) = lines.next_line().await {
                if line.is_empty() {
                    continue;
                }
                if let Ok(chunk) = serde_json::from_str::<OllamaChatChunk>(&line) {
                    let done = chunk.done;
                    if tx.send(chunk).await.is_err() {
                        return;
                    }
                    if done {
                        return;
                    }
                }
            }
        });

        Ok(rx)
    }

    pub async fn extract_entities(
        &self,
        text: &str,
    ) -> Result<(Vec<ExtractedEntity>, Vec<ExtractedRelation>)> {
        let prompt = format!(
            r#"Extract entities and relationships from this text. Return ONLY valid JSON with no extra text:
{{"entities": [{{"name": "...", "type": "...", "summary": "..."}}],
 "relations": [{{"source": "...", "target": "...", "type": "...", "fact": "..."}}]}}

Text: {text}"#
        );

        let raw = self.generate(&prompt).await?;

        // Try to find JSON in the response
        let json_str = if let Some(start) = raw.find('{') {
            if let Some(end) = raw.rfind('}') {
                &raw[start..=end]
            } else {
                &raw
            }
        } else {
            &raw
        };

        let data: ExtractedData =
            serde_json::from_str(json_str).context(format!("parse extraction response: {raw}"))?;
        Ok((data.entities, data.relations))
    }
}
