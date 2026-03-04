"use client";

import { motion } from "framer-motion";
import type { Entity } from "@/app/lib/types";

const typeColors: Record<string, string> = {
  entity: "border-indigo-500/30 bg-indigo-500/5",
  person: "border-blue-500/30 bg-blue-500/5",
  organization: "border-purple-500/30 bg-purple-500/5",
  concept: "border-green-500/30 bg-green-500/5",
  location: "border-orange-500/30 bg-orange-500/5",
};

export function EntityCard({ entity }: { entity: Entity }) {
  const color = typeColors[entity.type?.toLowerCase()] || typeColors.entity;

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className={`rounded-lg border p-2.5 ${color}`}
    >
      <div className="flex items-start justify-between gap-2">
        <h4 className="text-sm font-medium truncate">{entity.name}</h4>
        <span className="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
          {entity.type}
        </span>
      </div>
      {entity.summary && (
        <p className="mt-1 text-xs text-muted-foreground line-clamp-2">
          {entity.summary}
        </p>
      )}
    </motion.div>
  );
}
