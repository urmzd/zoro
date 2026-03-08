use anyhow::{Context, Result};
use reqwest::Client;
use serde::Deserialize;
use std::collections::HashSet;
use std::time::Duration;

use crate::models::SearchResult;

#[derive(Deserialize)]
struct SearxngResponse {
    results: Vec<SearxngResult>,
}

#[derive(Deserialize)]
struct SearxngResult {
    title: String,
    url: String,
    content: Option<String>,
}

pub struct Searcher {
    http: Client,
}

const SEARXNG_URL: &str = "http://127.0.0.1:8888";

impl Default for Searcher {
    fn default() -> Self {
        Self::new()
    }
}

impl Searcher {
    pub fn new() -> Self {
        Self {
            http: Client::builder()
                .timeout(Duration::from_secs(15))
                .build()
                .expect("http client"),
        }
    }

    /// Search using SearXNG JSON API (multi-engine aggregation).
    pub async fn search(&self, query: &str) -> Result<Vec<SearchResult>> {
        let resp = self
            .http
            .get(format!("{}/search", SEARXNG_URL))
            .query(&[
                ("q", query),
                ("format", "json"),
                ("engines", "google,bing"),
            ])
            .send()
            .await
            .context("searxng search request")?;

        if !resp.status().is_success() {
            anyhow::bail!("searxng returned {}", resp.status());
        }

        let body: SearxngResponse = resp.json().await.context("parse searxng response")?;

        let mut results = Vec::new();
        let mut seen_urls = HashSet::new();

        for r in body.results {
            if r.url.is_empty() || seen_urls.contains(&r.url) {
                continue;
            }
            seen_urls.insert(r.url.clone());

            results.push(SearchResult {
                title: r.title,
                url: r.url,
                snippet: r.content.unwrap_or_default(),
            });

            if results.len() >= 8 {
                break;
            }
        }

        log::info!("searxng returned {} results for {:?}", results.len(), query);
        Ok(results)
    }
}
