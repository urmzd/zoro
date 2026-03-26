import type { ChatSession, GraphData, NodeDetail, SearchFactsResponse } from "./types";

const API_BASE = "/api";

export interface ChatSessionSummary {
  id: string;
  preview: string;
  message_count: number;
  created_at: string;
}

// ── Chat ────────────────────────────────────────────────────────────

export async function createChatSession(): Promise<ChatSession> {
  const resp = await fetch(`${API_BASE}/sessions`, { method: "POST" });
  if (!resp.ok) throw new Error(`createChatSession: ${resp.status}`);
  return resp.json();
}

export async function listChatSessions(): Promise<ChatSessionSummary[]> {
  try {
    const resp = await fetch(`${API_BASE}/sessions`);
    if (!resp.ok) return [];
    return resp.json();
  } catch {
    return [];
  }
}

export async function getChatSession(sessionId: string): Promise<ChatSession> {
  const resp = await fetch(`${API_BASE}/sessions/${sessionId}`);
  if (!resp.ok) throw new Error(`getChatSession: ${resp.status}`);
  return resp.json();
}

export async function searchSessions(query: string): Promise<ChatSessionSummary[]> {
  try {
    const resp = await fetch(`${API_BASE}/sessions/search?q=${encodeURIComponent(query)}`);
    if (!resp.ok) return [];
    return resp.json();
  } catch {
    return [];
  }
}

// ── Knowledge ───────────────────────────────────────────────────────

export async function searchKnowledge(query: string): Promise<SearchFactsResponse> {
  const resp = await fetch(`${API_BASE}/knowledge/search?q=${encodeURIComponent(query)}`);
  if (!resp.ok) throw new Error(`searchKnowledge: ${resp.status}`);
  return resp.json();
}

export async function getKnowledgeGraph(limit = 300): Promise<GraphData> {
  const resp = await fetch(`${API_BASE}/knowledge/graph?limit=${limit}`);
  if (!resp.ok) throw new Error(`getKnowledgeGraph: ${resp.status}`);
  return resp.json();
}

export async function getNodeDetail(id: string, depth = 1): Promise<NodeDetail> {
  const resp = await fetch(`${API_BASE}/knowledge/nodes/${id}?depth=${depth}`);
  if (!resp.ok) throw new Error(`getNodeDetail: ${resp.status}`);
  return resp.json();
}

// ── Intent & Autocomplete ───────────────────────────────────────────

export async function classifyIntent(
  query: string,
): Promise<{ action: "chat" | "knowledge_search"; query: string }> {
  try {
    const resp = await fetch(`${API_BASE}/intent/classify`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query }),
    });
    if (!resp.ok) return { action: "chat", query };
    return resp.json();
  } catch {
    return { action: "chat", query };
  }
}

export async function getAutocompleteSuggestions(
  q: string,
  signal?: AbortSignal,
): Promise<string[]> {
  try {
    const resp = await fetch(`${API_BASE}/autocomplete?q=${encodeURIComponent(q)}`, { signal });
    if (!resp.ok) return [];
    const data = await resp.json();
    return data.suggestions ?? [];
  } catch {
    return [];
  }
}

// ── Status & Logs ──────────────────────────────────────────────────

export interface ServiceStatus {
  surrealdb: boolean;
  searxng: boolean;
  ollama: boolean;
}

export async function getStatus(): Promise<ServiceStatus> {
  const resp = await fetch(`${API_BASE}/status`);
  if (!resp.ok) return { surrealdb: false, searxng: false, ollama: false };
  return resp.json();
}

export async function getLogs(lines = 200): Promise<string> {
  try {
    const resp = await fetch(`${API_BASE}/logs?lines=${lines}`);
    if (!resp.ok) return "";
    const data = await resp.json();
    return data.logs ?? "";
  } catch {
    return "";
  }
}

// ── SSE helpers ─────────────────────────────────────────────────────

export interface SSEEvent {
  type: string;
  data: unknown;
}

export function sendMessageSSE(
  sessionId: string,
  content: string,
  onEvent: (event: SSEEvent) => void,
  onError?: (error: Error) => void,
): () => void {
  const abortController = new AbortController();

  fetch(`${API_BASE}/sessions/${sessionId}/messages`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ content }),
    signal: abortController.signal,
  })
    .then(async (resp) => {
      if (!resp.ok || !resp.body) {
        onError?.(new Error(`sendMessage: ${resp.status}`));
        return;
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            try {
              const event = JSON.parse(line.slice(6)) as SSEEvent;
              onEvent(event);
            } catch {
              // skip malformed lines
            }
          }
        }
      }
    })
    .catch((err) => {
      if (err.name !== "AbortError") {
        onError?.(err);
      }
    });

  return () => abortController.abort();
}

export function startResearchSSE(
  query: string,
  onEvent: (event: SSEEvent) => void,
  onError?: (error: Error) => void,
): () => void {
  const abortController = new AbortController();

  fetch(`${API_BASE}/research`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query }),
    signal: abortController.signal,
  })
    .then(async (resp) => {
      if (!resp.ok || !resp.body) {
        onError?.(new Error(`startResearch: ${resp.status}`));
        return;
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            try {
              const event = JSON.parse(line.slice(6)) as SSEEvent;
              onEvent(event);
            } catch {
              // skip malformed lines
            }
          }
        }
      }
    })
    .catch((err) => {
      if (err.name !== "AbortError") {
        onError?.(err);
      }
    });

  return () => abortController.abort();
}
