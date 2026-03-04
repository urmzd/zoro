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
