"use client";

import { useCallback, useEffect, useRef } from "react";
import dynamic from "next/dynamic";
import { useKnowledgeStore } from "@/lib/stores/knowledge-store";
import { getKnowledgeGraph, getNodeDetail, searchKnowledge } from "@/app/lib/api";
import { NodeDetailPanel } from "@/components/graph/node-detail-panel";
import { GraphControls } from "@/components/graph/graph-controls";
import { IconSearch } from "@tabler/icons-react";

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
  const fgRef = useRef<any>(null);
  const {
    graphData,
    selectedNode,
    searchQuery,
    isLoading,
    highlightedNodes,
    setGraphData,
    setSelectedNode,
    setSearchQuery,
    setLoading,
    highlightSubgraph,
  } = useKnowledgeStore();

  useEffect(() => {
    loadGraph();
  }, []);

  async function loadGraph() {
    setLoading(true);
    try {
      const data = await getKnowledgeGraph(300);
      setGraphData(data);
    } catch (err) {
      console.error("Failed to load graph:", err);
    } finally {
      setLoading(false);
    }
  }

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!searchQuery.trim()) {
      highlightSubgraph([]);
      return;
    }
    try {
      const results = await searchKnowledge(searchQuery);
      const nodeIds = new Set<string>();
      if (results?.facts) {
        for (const fact of results.facts) {
          if (fact.source_node?.uuid) nodeIds.add(fact.source_node.uuid);
          if (fact.target_node?.uuid) nodeIds.add(fact.target_node.uuid);
        }
      }
      highlightSubgraph(Array.from(nodeIds));
    } catch (err) {
      console.error("Search failed:", err);
    }
  }

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
    (node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
      const label = node.name || node.id;
      const fontSize = 12 / globalScale;
      const baseColor = NODE_COLORS[node.type] || NODE_COLORS.entity;
      const isHighlighted =
        highlightedNodes.size === 0 || highlightedNodes.has(node.id);
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
    [highlightedNodes]
  );

  return (
    <div className="relative h-full w-full">
      {/* Search overlay */}
      <form
        onSubmit={handleSearch}
        className="absolute top-4 left-4 z-10 flex items-center gap-2 rounded-lg border border-border/50 bg-background/80 backdrop-blur-md px-3 py-2"
      >
        <IconSearch className="h-4 w-4 text-muted-foreground" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search knowledge..."
          className="w-64 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
        />
      </form>

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
        <NodeDetailPanel
          detail={selectedNode}
          onClose={() => setSelectedNode(null)}
        />
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
