"use client";

import { useEffect, useRef } from "react";
import { createResearchStream } from "./api";
import { useResearchStore } from "@/lib/stores/research-store";

export function useResearchStream(sessionId: string | null, query: string) {
  const store = useResearchStore();
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!sessionId) return;

    store.connect(query);

    const es = createResearchStream(sessionId);
    esRef.current = es;

    es.addEventListener("search_started", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      store.setStatus("running");
      store.addTimelineEvent("search_started", data.query || "Searching...");
    });

    es.addEventListener("search_results", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      store.setSearchResults(Array.isArray(data) ? data : []);
      store.addTimelineEvent("search_results", `Found ${Array.isArray(data) ? data.length : 0} results`);
    });

    es.addEventListener("prior_knowledge", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      if (Array.isArray(data)) {
        store.setPriorFacts(data);
        store.addTimelineEvent("prior_knowledge", `Found ${data.length} related facts from prior research`);
      } else {
        store.addTimelineEvent("prior_knowledge", data.message || "Checking prior knowledge...");
      }
    });

    es.addEventListener("episode_ingested", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      store.addTimelineEvent("episode_ingested", `Ingested result ${(data.result_index ?? 0) + 1}`);
    });

    es.addEventListener("entity_discovered", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      store.addEntity(data);
      store.addTimelineEvent("entity_discovered", `Discovered: ${data.name}`);
    });

    es.addEventListener("relation_found", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      store.addRelation(data);
      store.addTimelineEvent("relation_found", `Relation: ${data.type}`);
    });

    es.addEventListener("graph_ready", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      if (data.facts) {
        const nodes = new Map<string, { id: string; name: string; type: string; summary: string }>();
        const edges: { id: string; source: string; target: string; type: string; fact: string; weight: number }[] = [];
        for (const fact of data.facts) {
          if (fact.source_node) {
            nodes.set(fact.source_node.uuid, {
              id: fact.source_node.uuid,
              name: fact.source_node.name,
              type: fact.source_node.type || "entity",
              summary: fact.source_node.summary || "",
            });
          }
          if (fact.target_node) {
            nodes.set(fact.target_node.uuid, {
              id: fact.target_node.uuid,
              name: fact.target_node.name,
              type: fact.target_node.type || "entity",
              summary: fact.target_node.summary || "",
            });
          }
          edges.push({
            id: fact.uuid,
            source: fact.source_node?.uuid,
            target: fact.target_node?.uuid,
            type: fact.name,
            fact: fact.fact,
            weight: 1,
          });
        }
        store.setGraphData({
          nodes: Array.from(nodes.values()),
          edges,
        });
      }
      store.addTimelineEvent("graph_ready", "Knowledge graph ready");
    });

    es.addEventListener("summary_token", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      store.appendSummary(data.token || "");
    });

    es.addEventListener("research_complete", () => {
      store.setStatus("complete");
      store.addTimelineEvent("research_complete", "Research complete");
      es.close();
    });

    es.addEventListener("error", (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data);
        store.setError(data.error || "Unknown error");
      } catch {
        store.setError("Connection error");
      }
      es.close();
    });

    es.onerror = () => {
      if (es.readyState === EventSource.CLOSED) return;
      store.setError("Connection lost");
      es.close();
    };

    return () => {
      es.close();
      esRef.current = null;
    };
  }, [sessionId, query]);

  return store;
}
