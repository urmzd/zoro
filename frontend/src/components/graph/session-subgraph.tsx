"use client";

import dynamic from "next/dynamic";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Entity, GraphData, Relation } from "@/app/lib/types";

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

interface SessionSubgraphProps {
  entities: Entity[];
  relations: Relation[];
  graphData: GraphData | null;
}

export function SessionSubgraph({ entities, relations, graphData }: SessionSubgraphProps) {
  // biome-ignore lint/suspicious/noExplicitAny: react-force-graph-2d ref type
  const fgRef = useRef<any>(null);
  const [newNodeIds, setNewNodeIds] = useState<Set<string>>(new Set());
  const prevCountRef = useRef(0);

  // Track newly added nodes for pulse animation
  useEffect(() => {
    const currentCount = entities.length + (graphData?.nodes.length || 0);
    if (currentCount > prevCountRef.current) {
      const allIds = graphData?.nodes.map((n) => n.id) || entities.map((e) => e.uuid);
      const newIds = new Set(allIds.slice(prevCountRef.current));
      setNewNodeIds(newIds);
      const timer = setTimeout(() => setNewNodeIds(new Set()), 2000);
      prevCountRef.current = currentCount;
      return () => clearTimeout(timer);
    }
  }, [entities, graphData]);

  const data = useMemo(() => {
    if (graphData && graphData.nodes.length > 0) {
      return {
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
          fact: e.fact,
        })),
      };
    }

    // Build from entities/relations if no graphData yet
    if (entities.length === 0) return { nodes: [], links: [] };

    // biome-ignore lint/suspicious/noExplicitAny: graph node shape is dynamic
    const nodeMap = new Map<string, any>();
    for (const e of entities) {
      nodeMap.set(e.uuid, {
        id: e.uuid,
        name: e.name,
        type: e.type || "entity",
        val: 3,
      });
    }

    const links = relations
      .filter((r) => nodeMap.has(r.source_uuid) && nodeMap.has(r.target_uuid))
      .map((r) => ({
        source: r.source_uuid,
        target: r.target_uuid,
        type: r.type,
        fact: r.fact,
      }));

    return { nodes: Array.from(nodeMap.values()), links };
  }, [entities, relations, graphData]);

  const nodeCanvasObject = useCallback(
    // biome-ignore lint/suspicious/noExplicitAny: react-force-graph-2d callback type
    (node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
      const label = node.name || node.id;
      const fontSize = 12 / globalScale;
      const color = NODE_COLORS[node.type] || NODE_COLORS.entity;
      const isNew = newNodeIds.has(node.id);
      const r = isNew ? 7 : 5;

      // Pulse ring for new nodes
      if (isNew) {
        const pulseR = r + 4 + Math.sin(Date.now() / 200) * 3;
        ctx.beginPath();
        ctx.arc(node.x, node.y, pulseR, 0, 2 * Math.PI, false);
        ctx.strokeStyle = color;
        ctx.lineWidth = 1;
        ctx.globalAlpha = 0.4;
        ctx.stroke();
        ctx.globalAlpha = 1;
      }

      // Node circle
      ctx.beginPath();
      ctx.arc(node.x, node.y, r, 0, 2 * Math.PI, false);
      ctx.fillStyle = color;
      ctx.fill();

      // Glow
      ctx.shadowColor = color;
      ctx.shadowBlur = isNew ? 16 : 8;
      ctx.fill();
      ctx.shadowBlur = 0;

      // Label
      ctx.font = `${fontSize}px sans-serif`;
      ctx.textAlign = "center";
      ctx.textBaseline = "top";
      ctx.fillStyle = "rgba(255,255,255,0.8)";
      ctx.fillText(label, node.x, node.y + r + 2);
    },
    [newNodeIds],
  );

  // biome-ignore lint/suspicious/noExplicitAny: react-force-graph-2d callback type
  const linkCanvasObject = useCallback((link: any, ctx: CanvasRenderingContext2D) => {
    const start = link.source;
    const end = link.target;
    if (!start || !end || typeof start.x !== "number") return;

    ctx.beginPath();
    ctx.moveTo(start.x, start.y);
    ctx.lineTo(end.x, end.y);
    ctx.strokeStyle = "rgba(99, 102, 241, 0.25)";
    ctx.lineWidth = 1;
    ctx.stroke();
  }, []);

  if (data.nodes.length === 0) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Graph will appear as entities are discovered...
      </div>
    );
  }

  return (
    <ForceGraph2D
      ref={fgRef}
      graphData={data}
      nodeCanvasObject={nodeCanvasObject}
      linkCanvasObject={linkCanvasObject}
      backgroundColor="transparent"
      cooldownTicks={100}
      width={undefined}
      height={undefined}
    />
  );
}
