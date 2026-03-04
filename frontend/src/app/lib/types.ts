export interface SearchResult {
  title: string;
  url: string;
  snippet: string;
}

export interface Entity {
  uuid: string;
  name: string;
  type: string;
  summary: string;
}

export interface Relation {
  uuid: string;
  source_uuid: string;
  target_uuid: string;
  type: string;
  fact: string;
}

export interface Fact {
  uuid: string;
  name: string;
  fact: string;
  source_node: Entity;
  target_node: Entity;
}

export interface TimelineEvent {
  type: SSEEventType;
  message: string;
  timestamp: string;
  data?: Record<string, unknown>;
}

export type SSEEventType =
  | "search_started"
  | "search_results"
  | "episode_ingested"
  | "entity_discovered"
  | "relation_found"
  | "prior_knowledge"
  | "graph_ready"
  | "summary_token"
  | "research_complete"
  | "error";

export interface GraphNode {
  id: string;
  name: string;
  type: string;
  summary?: string;
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  type: string;
  fact?: string;
  weight: number;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export interface NodeDetail {
  node: GraphNode;
  neighbors: GraphNode[];
  edges: GraphEdge[];
}

export interface ResearchSession {
  id: string;
  query: string;
  status: string;
  results: SearchResult[];
  entities: Entity[];
  relations: Relation[];
  timeline: TimelineEvent[];
  summary: string;
  created_at: string;
}

export interface ResearchState {
  status: "idle" | "connecting" | "running" | "complete" | "error";
  query: string;
  searchResults: SearchResult[];
  entities: Entity[];
  relations: Relation[];
  priorFacts: Fact[];
  graphData: GraphData | null;
  summary: string;
  timeline: TimelineEvent[];
  error: string | null;
}
