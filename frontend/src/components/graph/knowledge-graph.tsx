"use client";

import dynamic from "next/dynamic";
import { useCallback, useEffect, useRef } from "react";
import { getKnowledgeGraph, getNodeDetail } from "@/app/lib/api";
import { GraphControls } from "@/components/graph/graph-controls";
import { NodeDetailPanel } from "@/components/graph/node-detail-panel";
import { useKnowledgeStore } from "@/lib/stores/knowledge-store";

const ForceGraph2D = dynamic(() => import("react-force-graph-2d"), {
  ssr: false,
});

const NODE_COLORS: Record<string, string> = {
  entity: "#6366f1",
  person: "#3b82f6",
  organization: "#a855f7",
  concept: "#22c55e",
  location: "#f97316",
  episode: "#22c55e",
  community: "#a855f7",
  source: "#f97316",
};

export function KnowledgeGraph() {
  // biome-ignore lint/suspicious/noExplicitAny: react-force-graph-2d ref has no exported type
  const fgRef = useRef<any>(null);
  const {
    graphData,
    selectedNode,
    isLoading,
    highlightedNodes,
    setGraphData,
    setSelectedNode,
    setLoading,
  } = useKnowledgeStore();

  const loadGraph = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getKnowledgeGraph(300);
      setGraphData(data);
    } catch (err) {
      console.error("Failed to load graph:", err);
    } finally {
      setLoading(false);
    }
  }, [setLoading, setGraphData]);

  useEffect(() => {
    loadGraph();
  }, [loadGraph]);

  // biome-ignore lint/suspicious/noExplicitAny: react-force-graph-2d node type
  async function handleNodeClick(node: any) {
    try {
      const detail = await getNodeDetail(node.id);
      setSelectedNode(detail);
    } catch {
      setSelectedNode({
        node: { id: node.id, name: node.name, type: node.type, summary: "" },
        neighbors: [],
        edges: [],
      });
    }
  }

  const data = graphData
    ? {
        nodes: graphData.nodes.map((n) => ({
          id: n.id,
          name: n.name,
          type: n.type || "entity",
          val: 3,
        })),
        links: graphData.edges.map((e) => ({
          source: e.source,
          target: e.target,
          type: e.type,
        })),
      }
    : { nodes: [], links: [] };

  const nodeCanvasObject = useCallback(
    // biome-ignore lint/suspicious/noExplicitAny: react-force-graph-2d callback type
    (node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
      const label = node.name || node.id;
      const fontSize = 12 / globalScale;
      const baseColor = NODE_COLORS[node.type] || NODE_COLORS.entity;
      const isHighlighted = highlightedNodes.size === 0 || highlightedNodes.has(node.id);
      const color = isHighlighted ? baseColor : "rgba(100,100,100,0.3)";
      const r = isHighlighted ? 6 : 4;

      ctx.beginPath();
      ctx.arc(node.x, node.y, r, 0, 2 * Math.PI, false);
      ctx.fillStyle = color;
      ctx.fill();

      if (isHighlighted && highlightedNodes.size > 0) {
        ctx.shadowColor = baseColor;
        ctx.shadowBlur = 12;
        ctx.fill();
        ctx.shadowBlur = 0;
      }

      if (isHighlighted) {
        ctx.font = `${fontSize}px sans-serif`;
        ctx.textAlign = "center";
        ctx.textBaseline = "top";
        ctx.fillStyle = "rgba(255,255,255,0.8)";
        ctx.fillText(label, node.x, node.y + r + 2);
      }
    },
    [highlightedNodes],
  );

  return (
    <div className="relative h-full w-full">
      {/* Stats overlay */}
      <div className="absolute top-4 right-4 z-10 rounded-lg border border-border/50 bg-background/80 backdrop-blur-md px-3 py-2 text-xs text-muted-foreground">
        {data.nodes.length} nodes / {data.links.length} edges
      </div>

      {/* Graph */}
      {isLoading ? (
        <div className="flex h-full items-center justify-center text-muted-foreground">
          Loading knowledge graph...
        </div>
      ) : data.nodes.length === 0 ? (
        <div className="flex h-full items-center justify-center text-muted-foreground">
          No knowledge yet. Run some research first!
        </div>
      ) : (
        <ForceGraph2D
          ref={fgRef}
          graphData={data}
          nodeCanvasObject={nodeCanvasObject}
          linkColor={() => "rgba(99, 102, 241, 0.2)"}
          linkWidth={1}
          backgroundColor="transparent"
          onNodeClick={handleNodeClick}
          cooldownTicks={200}
        />
      )}

      {/* Node detail panel */}
      {selectedNode && (
        <NodeDetailPanel detail={selectedNode} onClose={() => setSelectedNode(null)} />
      )}

      {/* Graph controls */}
      <GraphControls
        onZoomIn={() => fgRef.current?.zoom(2, 400)}
        onZoomOut={() => fgRef.current?.zoom(0.5, 400)}
        onReset={() => fgRef.current?.zoomToFit(400)}
      />
    </div>
  );
}
