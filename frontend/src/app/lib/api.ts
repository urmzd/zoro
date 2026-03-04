import { ApiClient } from "@/generated/api";
import type { GraphData, NodeDetail } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const apiClient = new ApiClient({
  baseUrl: API_BASE,
  retry: false,
});

export async function startResearch(query: string): Promise<string> {
  const data = await apiClient.startResearch({ query });
  return data.id!;
}

export function createResearchStream(sessionId: string): EventSource {
  return new EventSource(`${API_BASE}/api/research/${sessionId}/stream`);
}

export async function getResearch(sessionId: string) {
  return apiClient.getResearchSession(sessionId);
}

export async function searchKnowledge(query: string) {
  return apiClient.searchKnowledge(query);
}

export async function getKnowledgeGraph(limit = 300): Promise<GraphData> {
  return apiClient.getKnowledgeGraph(limit) as Promise<GraphData>;
}

export async function getNodeDetail(id: string, depth = 1): Promise<NodeDetail> {
  return apiClient.getNodeDetail(id, depth) as Promise<NodeDetail>;
}

// Intent & Autocomplete API functions

export async function classifyIntent(
  query: string,
): Promise<{ action: "chat" | "knowledge_search"; query: string }> {
  const resp = await fetch(`${API_BASE}/api/intent`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query }),
  });
  if (!resp.ok) return { action: "chat", query };
  return resp.json();
}

export async function getAutocompleteSuggestions(
  q: string,
  signal?: AbortSignal,
): Promise<string[]> {
  const resp = await fetch(
    `${API_BASE}/api/autocomplete?q=${encodeURIComponent(q)}`,
    { signal },
  );
  if (!resp.ok) return [];
  const data = await resp.json();
  return data.suggestions ?? [];
}

// Chat API functions

export async function createChatSession(): Promise<{ id: string }> {
  const resp = await fetch(`${API_BASE}/api/chat/sessions`, {
    method: "POST",
  });
  if (!resp.ok) throw new Error("Failed to create chat session");
  return resp.json();
}

export async function getChatSession(sessionId: string) {
  const resp = await fetch(`${API_BASE}/api/chat/sessions/${sessionId}`);
  if (!resp.ok) throw new Error("Failed to get chat session");
  return resp.json();
}

export async function sendChatMessage(
  sessionId: string,
  content: string,
): Promise<Response> {
  return fetch(`${API_BASE}/api/chat/sessions/${sessionId}/messages`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ content }),
  });
}
