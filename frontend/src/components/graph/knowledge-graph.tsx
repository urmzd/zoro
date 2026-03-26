"use client";

import { useCallback, useEffect, useRef } from "react";
import { getKnowledgeGraph, getNodeDetail } from "@/app/lib/api";
import { GraphControls } from "@/components/graph/graph-controls";
import { LazyForceGraph2D as ForceGraph2D } from "@/components/graph/lazy-force-graph";
import { NodeDetailPanel } from "@/components/graph/node-detail-panel";
import { useKnowledgeStore } from "@/lib/stores/knowledge-store";

const NODE_COLORS: Record<string, string> = {
  entity: "#ba9eff",
  person: "#699cff",
  organization: "#a27cff",
  concept: "#22c55e",
  location: "#ff716a",
  episode: "#22c55e",
  community: "#a27cff",
  source: "#ff716a",
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

  const hasFetched = useRef(false);

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
    if (hasFetched.current) return;
    hasFetched.current = true;
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
        nodes: (graphData.nodes ?? []).map((n) => ({
          id: n.id,
          name: n.name,
          type: n.type || "entity",
          val: 3,
        })),
        links: (graphData.edges ?? []).map((e) => ({
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
        ctx.fillStyle = "rgba(222,229,255,0.8)";
        ctx.fillText(label, node.x, node.y + r + 2);
      }
    },
    [highlightedNodes],
  );

  return (
    <div className="relative h-full w-full">
      {/* Stats overlay */}
      <div className="absolute top-6 left-6 z-10 p-4 bg-[#050a18]/60 backdrop-blur-md rounded-xl border border-[#40485d]/10">
        <div className="flex items-center gap-6">
          <div className="flex flex-col">
            <span className="text-[10px] text-[#a3aac4] font-bold uppercase tracking-widest">
              Nodes
            </span>
            <span className="text-xl font-headline font-bold">
              {data.nodes.length.toLocaleString()}
            </span>
          </div>
          <div className="flex flex-col">
            <span className="text-[10px] text-[#a3aac4] font-bold uppercase tracking-widest">
              Edges
            </span>
            <span className="text-xl font-headline font-bold">
              {data.links.length.toLocaleString()}
            </span>
          </div>
        </div>
      </div>

      {/* Graph */}
      {isLoading ? (
        <div className="flex h-full items-center justify-center text-[#a3aac4]">
          <span className="material-symbols-outlined text-3xl mr-3 animate-spin">
            progress_activity
          </span>
          Loading knowledge graph...
        </div>
      ) : data.nodes.length === 0 ? (
        <div className="flex h-full flex-col items-center justify-center text-[#a3aac4]">
          <span className="material-symbols-outlined text-5xl mb-4 opacity-30">hub</span>
          <p>No knowledge yet. Run some research first!</p>
        </div>
      ) : (
        <ForceGraph2D
          ref={fgRef}
          graphData={data}
          nodeCanvasObject={nodeCanvasObject}
          linkColor={() => "rgba(186, 158, 255, 0.15)"}
          linkWidth={1}
          linkDirectionalArrowLength={4}
          linkDirectionalArrowRelPos={1}
          linkDirectionalArrowColor={() => "rgba(186, 158, 255, 0.3)"}
          linkCanvasObjectMode={() => "after"}
          // biome-ignore lint/suspicious/noExplicitAny: react-force-graph-2d has no exported link types
          linkCanvasObject={(link: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
            // biome-ignore lint/suspicious/noExplicitAny: force-graph resolves node refs at runtime
            const source = link.source as any;
            // biome-ignore lint/suspicious/noExplicitAny: force-graph resolves node refs at runtime
            const target = link.target as any;
            if (!source?.x || !target?.x) return;
            const label = link.type || "";
            if (!label) return;

            const midX = (source.x + target.x) / 2;
            const midY = (source.y + target.y) / 2;
            const fontSize = Math.max(10 / globalScale, 1);

            ctx.font = `${fontSize}px sans-serif`;
            ctx.textAlign = "center";
            ctx.textBaseline = "middle";
            ctx.fillStyle = "rgba(163, 170, 196, 0.5)";
            ctx.fillText(label, midX, midY);
          }}
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

      {/* Legend */}
      <div className="absolute bottom-6 right-6 z-10 p-4 bg-[#050a18]/60 backdrop-blur-md rounded-xl border border-[#40485d]/10">
        <div className="flex gap-4">
          {Object.entries(NODE_COLORS)
            .slice(0, 4)
            .map(([type, color]) => (
              <div key={type} className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full" style={{ backgroundColor: color }} />
                <span className="text-[10px] font-bold text-[#a3aac4] uppercase">{type}</span>
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}
