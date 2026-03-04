"use client";

import { motion } from "framer-motion";
import type { NodeDetail } from "@/app/lib/types";
import { IconX } from "@tabler/icons-react";

interface NodeDetailPanelProps {
  detail: NodeDetail;
  onClose: () => void;
}

export function NodeDetailPanel({ detail, onClose }: NodeDetailPanelProps) {
  return (
    <motion.div
      initial={{ x: "100%" }}
      animate={{ x: 0 }}
      exit={{ x: "100%" }}
      transition={{ type: "spring", damping: 25, stiffness: 300 }}
      className="absolute top-0 right-0 z-20 h-full w-80 border-l border-border bg-background/95 backdrop-blur-md overflow-y-auto"
    >
      <div className="p-4">
        <div className="flex items-start justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold">{detail.node.name}</h3>
            <span className="inline-block rounded-full bg-indigo-500/20 text-indigo-400 px-2 py-0.5 text-xs mt-1">
              {detail.node.type}
            </span>
          </div>
          <button
            onClick={onClose}
            className="rounded-md p-1 hover:bg-muted transition-colors"
          >
            <IconX className="h-4 w-4" />
          </button>
        </div>

        {detail.node.summary && (
          <div className="mb-4">
            <h4 className="text-sm font-medium text-muted-foreground mb-1">Summary</h4>
            <p className="text-sm">{detail.node.summary}</p>
          </div>
        )}

        {detail.edges.length > 0 && (
          <div className="mb-4">
            <h4 className="text-sm font-medium text-muted-foreground mb-2">
              Relations ({detail.edges.length})
            </h4>
            <ul className="space-y-2">
              {detail.edges.map((edge) => (
                <li
                  key={edge.id}
                  className="rounded-md border border-border/50 p-2 text-sm"
                >
                  <span className="text-indigo-400">{edge.type}</span>
                  {edge.fact && (
                    <p className="text-xs text-muted-foreground mt-1">
                      {edge.fact}
                    </p>
                  )}
                </li>
              ))}
            </ul>
          </div>
        )}

        {detail.neighbors.length > 0 && (
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-2">
              Connected Nodes ({detail.neighbors.length})
            </h4>
            <ul className="space-y-1">
              {detail.neighbors.map((n) => (
                <li
                  key={n.id}
                  className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted/50 transition-colors"
                >
                  <span className="h-2 w-2 rounded-full bg-indigo-500 shrink-0" />
                  <span className="truncate">{n.name}</span>
                  <span className="text-[10px] text-muted-foreground ml-auto">
                    {n.type}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </motion.div>
  );
}
