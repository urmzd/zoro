const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function startResearch(query: string): Promise<string> {
  const res = await fetch(`${API_BASE}/api/research`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query }),
  });
  if (!res.ok) throw new Error("Failed to start research");
  const data = await res.json();
  return data.id;
}

export function createResearchStream(sessionId: string): EventSource {
  return new EventSource(`${API_BASE}/api/research/${sessionId}/stream`);
}

export async function getResearch(sessionId: string) {
  const res = await fetch(`${API_BASE}/api/research/${sessionId}`);
  if (!res.ok) throw new Error("Failed to get research");
  return res.json();
}

export async function searchKnowledge(query: string) {
  const res = await fetch(`${API_BASE}/api/knowledge/search?q=${encodeURIComponent(query)}`);
  if (!res.ok) throw new Error("Knowledge search failed");
  return res.json();
}

export async function getKnowledgeGraph(limit = 300) {
  const res = await fetch(`${API_BASE}/api/knowledge/graph?limit=${limit}`);
  if (!res.ok) throw new Error("Graph fetch failed");
  return res.json();
}

export async function getNodeDetail(id: string, depth = 1) {
  const res = await fetch(`${API_BASE}/api/knowledge/node/${id}?depth=${depth}`);
  if (!res.ok) throw new Error("Node fetch failed");
  return res.json();
}
