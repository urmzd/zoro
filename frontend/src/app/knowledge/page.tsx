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

  useEffect(() => {
    return () => resetChat();
  }, [resetChat]);

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <SearchSection onSearch={send} isStreaming={isStreaming} />

      <div className="flex flex-1 min-h-0 gap-6 px-8 pb-8">
        <div className="w-1/3">
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
        <div className="flex-1 relative bg-black/30 rounded-2xl border border-[#40485d]/10 overflow-hidden">
          {/* Grid pattern background */}
          <div className="absolute inset-0 bg-[radial-gradient(#1f2b49_1px,transparent_1px)] [background-size:32px_32px] opacity-40" />
          <KnowledgeGraph />
        </div>
      </div>
    </div>
  );
}
