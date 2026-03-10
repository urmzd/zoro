"use client";

import { useCallback, useEffect } from "react";
import { getKnowledgeGraph } from "@/app/lib/api";
import { useKnowledgeChatStream } from "@/app/lib/use-knowledge-chat-stream";
import { KnowledgeGraph } from "@/components/graph/knowledge-graph";
import { SearchSection } from "@/components/knowledge/search-section";
import { TraversalPanel } from "@/components/knowledge/traversal-panel";
import { useKnowledgeChatStore } from "@/lib/stores/knowledge-chat-store";
import { useKnowledgeStore } from "@/lib/stores/knowledge-store";

export default function KnowledgePage() {
  const knowledgeStore = useKnowledgeStore();
  const resetChat = useKnowledgeChatStore((s) => s.reset);

  const handleToolCallResult = useCallback(
    (name: string, result: string) => {
      if (name === "search_knowledge") {
        // Parse "- SourceName -> TargetName: fact" lines
        const nodeNames = new Set<string>();
        for (const line of result.split("\n")) {
          const match = line.match(/^-\s*(.+?)\s*->\s*(.+?):\s*/);
          if (match) {
            nodeNames.add(match[1].trim().toLowerCase());
            nodeNames.add(match[2].trim().toLowerCase());
          }
        }

        if (nodeNames.size > 0 && knowledgeStore.graphData) {
          const matchedIds = knowledgeStore.graphData.nodes
            .filter((n) => nodeNames.has(n.name.toLowerCase()))
            .map((n) => n.id);
          knowledgeStore.highlightSubgraph(matchedIds);
        }
      }

      if (name === "store_knowledge") {
        // Refresh graph when new knowledge is stored
        getKnowledgeGraph(300)
          .then((data) => knowledgeStore.setGraphData(data))
          .catch(() => {});
      }
    },
    [knowledgeStore],
  );

  const { messages, currentAssistantContent, currentToolCalls, status, error, send, stop } =
    useKnowledgeChatStream({ onToolCallResult: handleToolCallResult });

  const isStreaming = status === "streaming";

  // Reset chat store on unmount
  useEffect(() => {
    return () => resetChat();
  }, [resetChat]);

  return (
    <main className="flex h-screen flex-col overflow-hidden">
      {/* Prominent search section */}
      <SearchSection onSearch={send} isStreaming={isStreaming} />

      {/* Split panel: traversal | graph */}
      <div className="flex flex-1 min-h-0">
        {/* Left: Traversal / Chat panel */}
        <div className="w-[45%] border-r border-border/50">
          <TraversalPanel
            messages={messages}
            currentAssistantContent={currentAssistantContent}
            currentToolCalls={currentToolCalls}
            isStreaming={isStreaming}
            error={error}
            onSend={send}
            onStop={stop}
          />
        </div>

        {/* Right: Knowledge Graph */}
        <div className="w-[55%]">
          <KnowledgeGraph />
        </div>
      </div>
    </main>
  );
}
