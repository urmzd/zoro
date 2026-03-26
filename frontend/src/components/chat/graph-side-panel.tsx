"use client";

import { useCallback, useEffect, useState } from "react";
import { getKnowledgeGraph } from "@/app/lib/api";
import type { GraphData } from "@/app/lib/types";
import { SessionSubgraph } from "@/components/graph/session-subgraph";

interface GraphSidePanelProps {
  sessionId: string;
}

export function GraphSidePanel({ sessionId: _sessionId }: GraphSidePanelProps) {
  const [open, setOpen] = useState(false);
  const [graphData, setGraphData] = useState<GraphData | null>(null);

  const fetchGraph = useCallback(async () => {
    try {
      const data = await getKnowledgeGraph(200);
      setGraphData(data);
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    if (open) {
      fetchGraph();
    }
  }, [open, fetchGraph]);

  if (!open) {
    return (
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="fixed right-4 top-1/2 -translate-y-1/2 z-40 rounded-full border border-[#40485d]/20 bg-[#050a18]/80 backdrop-blur-sm p-3 text-[#a3aac4] hover:text-[#dee5ff] transition-colors shadow-lg"
        title="Open Knowledge Graph"
      >
        <span className="material-symbols-outlined">hub</span>
      </button>
    );
  }

  return (
    <div className="w-96 border-l border-[#40485d]/15 bg-[#060e20] flex flex-col shrink-0">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[#40485d]/10">
        <h3 className="text-sm font-medium font-headline">Knowledge Graph</h3>
        <button
          type="button"
          onClick={() => setOpen(false)}
          className="text-[#a3aac4] hover:text-[#dee5ff]"
        >
          <span className="material-symbols-outlined text-sm">close</span>
        </button>
      </div>
      <div className="flex-1 min-h-0">
        <SessionSubgraph entities={[]} relations={[]} graphData={graphData} />
      </div>
    </div>
  );
}
