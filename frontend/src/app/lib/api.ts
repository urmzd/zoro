import { invoke } from "@tauri-apps/api/core";
import type {
  ChatSession,
  GraphData,
  NodeDetail,
  SearchFactsResponse,
} from "./types";

export interface ChatSessionSummary {
  id: string;
  preview: string;
  message_count: number;
  created_at: string;
}

export async function startResearch(query: string): Promise<string> {
  return invoke<string>("start_research", { query });
}

export async function searchKnowledge(
  query: string,
): Promise<SearchFactsResponse> {
  return invoke<SearchFactsResponse>("search_knowledge", { query });
}

export async function getKnowledgeGraph(limit = 300): Promise<GraphData> {
  return invoke<GraphData>("get_knowledge_graph", { limit });
}

export async function getNodeDetail(
  id: string,
  depth = 1,
): Promise<NodeDetail> {
  return invoke<NodeDetail>("get_node_detail", { id, depth });
}

// Intent & Autocomplete

export async function classifyIntent(
  query: string,
): Promise<{ action: "chat" | "knowledge_search"; query: string }> {
  try {
    return await invoke<{ action: "chat" | "knowledge_search"; query: string }>(
      "classify_intent",
      { query },
    );
  } catch {
    return { action: "chat", query };
  }
}

export async function getAutocompleteSuggestions(
  q: string,
  _signal?: AbortSignal,
): Promise<string[]> {
  try {
    const resp = await invoke<{ suggestions: string[] }>("get_autocomplete", {
      query: q,
    });
    return resp.suggestions ?? [];
  } catch {
    return [];
  }
}

// Chat

export async function listChatSessions(): Promise<ChatSessionSummary[]> {
  try {
    return await invoke<ChatSessionSummary[]>("list_chat_sessions");
  } catch {
    return [];
  }
}

export async function createChatSession(): Promise<{ id: string }> {
  const session = await invoke<ChatSession>("create_chat_session");
  return { id: session.id };
}

export async function getChatSession(sessionId: string): Promise<ChatSession> {
  return invoke<ChatSession>("get_chat_session", { id: sessionId });
}

export async function sendChatMessage(
  sessionId: string,
  content: string,
): Promise<void> {
  return invoke<void>("send_chat_message", { id: sessionId, content });
}
