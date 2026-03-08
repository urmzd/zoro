"use client";

import { IconGraph, IconX } from "@tabler/icons-react";
import { useCallback, useEffect, useState } from "react";
import { getKnowledgeGraph } from "@/app/lib/api";
import type { GraphData } from "@/app/lib/types";
import { SessionSubgraph } from "@/components/graph/session-subgraph";

interface GraphSidePanelProps {
  sessionId: string;
}

export function GraphSidePanel({ sessionId }: GraphSidePanelProps) {
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
        className="fixed right-4 top-1/2 -translate-y-1/2 z-40 rounded-full border border-border bg-background/80 backdrop-blur-sm p-3 text-muted-foreground hover:text-foreground transition-colors shadow-lg"
        title="Open Knowledge Graph"
      >
        <IconGraph className="h-5 w-5" />
      </button>
    );
  }

  return (
    <div className="w-96 border-l border-border bg-background flex flex-col shrink-0">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border">
        <h3 className="text-sm font-medium">Knowledge Graph</h3>
        <button
          type="button"
          onClick={() => setOpen(false)}
          className="text-muted-foreground hover:text-foreground"
        >
          <IconX className="h-4 w-4" />
        </button>
      </div>
      <div className="flex-1 min-h-0">
        <SessionSubgraph entities={[]} relations={[]} graphData={graphData} />
      </div>
    </div>
  );
}
