"use client";

import { motion } from "framer-motion";
import type { SSEEventType, TimelineEvent } from "@/app/lib/types";

const eventStyles: Record<SSEEventType, { color: string; label: string }> = {
  search_started: { color: "bg-blue-500", label: "Search" },
  search_results: { color: "bg-blue-400", label: "Results" },
  episode_ingested: { color: "bg-amber-500", label: "Ingested" },
  entity_discovered: { color: "bg-green-500", label: "Entity" },
  relation_found: { color: "bg-purple-500", label: "Relation" },
  prior_knowledge: { color: "bg-indigo-500", label: "Prior" },
  graph_ready: { color: "bg-cyan-500", label: "Graph" },
  summary_token: { color: "bg-gray-500", label: "Summary" },
  research_complete: { color: "bg-emerald-500", label: "Complete" },
  error: { color: "bg-red-500", label: "Error" },
};

interface ResearchTimelineProps {
  events: TimelineEvent[];
}

export function ResearchTimeline({ events }: ResearchTimelineProps) {
  // Filter out summary_token events from timeline display
  const displayEvents = events.filter((e) => e.type !== "summary_token");

  if (displayEvents.length === 0) {
    return <div className="text-sm text-muted-foreground">Waiting for events...</div>;
  }

  return (
    <div className="space-y-2">
      {displayEvents.map((event, idx) => {
        const style = eventStyles[event.type] || eventStyles.error;
        return (
          <motion.div
            key={`${event.type}-${event.timestamp}`}
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: idx * 0.02 }}
            className="flex items-start gap-2"
          >
            <div className="flex flex-col items-center shrink-0 pt-1">
              <div className={`h-2 w-2 rounded-full ${style.color}`} />
              {idx < displayEvents.length - 1 && <div className="w-px h-full bg-border" />}
            </div>
            <div className="min-w-0 pb-2">
              <span className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">
                {style.label}
              </span>
              <p className="text-xs text-foreground/80 truncate">{event.message}</p>
            </div>
          </motion.div>
        );
      })}
    </div>
  );
}
