"use client";

import { useEffect, useRef } from "react";
import { useResearchStore } from "@/lib/stores/research-store";
import { type SSEEvent, startResearchSSE } from "./api";

export function useResearchStream(sessionId: string | null, query: string) {
  const store = useResearchStore();
  const cancelRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    if (!sessionId || !query) return;

    store.connect(query);

    const cancel = startResearchSSE(
      query,
      (event: SSEEvent) => {
        const { type, data } = event;
        // biome-ignore lint: data can be any shape from the backend
        const d = data as any;

        switch (type) {
          case "search_started":
            store.setStatus("running");
            store.addTimelineEvent("search_started", d?.query || "Searching...");
            break;

          case "search_results":
            store.setSearchResults(Array.isArray(d) ? d : []);
            store.addTimelineEvent(
              "search_results",
              `Found ${Array.isArray(d) ? d.length : 0} results`,
            );
            break;

          case "prior_knowledge":
            if (Array.isArray(d)) {
              store.setPriorFacts(d);
              store.addTimelineEvent(
                "prior_knowledge",
                `Found ${d.length} related facts from prior research`,
              );
            } else {
              store.addTimelineEvent(
                "prior_knowledge",
                d?.message || "Checking prior knowledge...",
              );
            }
            break;

          case "episode_ingested":
            store.addTimelineEvent(
              "episode_ingested",
              `Ingested result ${(d?.result_index ?? 0) + 1}`,
            );
            break;

          case "entity_discovered":
            store.addEntity(d);
            store.addTimelineEvent("entity_discovered", `Discovered: ${d?.name}`);
            break;

          case "relation_found":
            store.addRelation(d);
            store.addTimelineEvent("relation_found", `Relation: ${d?.type}`);
            break;

          case "graph_ready":
            if (d?.facts) {
              const nodes = new Map<
                string,
                { id: string; name: string; type: string; summary: string }
              >();
              const edges: {
                id: string;
                source: string;
                target: string;
                type: string;
                fact: string;
                weight: number;
              }[] = [];
              for (const fact of d.facts) {
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
            break;

          case "summary_token":
            store.appendSummary(d?.token || "");
            break;

          case "research_complete":
            store.setStatus("complete");
            store.addTimelineEvent("research_complete", "Research complete");
            break;

          case "error":
            store.setError(d?.error || "Unknown error");
            break;
        }
      },
      (err) => {
        store.setError(err.message);
      },
    );

    cancelRef.current = cancel;

    return () => {
      if (cancelRef.current) {
        cancelRef.current();
        cancelRef.current = null;
      }
    };
  }, [
    sessionId,
    query,
    store.addEntity,
    store.addRelation,
    store.addTimelineEvent,
    store.appendSummary,
    store.connect,
    store.setError,
    store.setGraphData,
    store.setPriorFacts,
    store.setSearchResults,
    store.setStatus,
  ]);

  return store;
}
