"use client";

import { useResearchStream } from "@/app/lib/use-research-stream";
import { ResearchTimeline } from "@/components/timeline/research-timeline";
import { SessionSubgraph } from "@/components/graph/session-subgraph";
import { EntityCard } from "@/components/research/entity-card";
import { StreamingSummary } from "@/components/research/streaming-summary";

interface ResearchViewProps {
  sessionId: string;
  query: string;
}

export function ResearchView({ sessionId, query }: ResearchViewProps) {
  const state = useResearchStream(sessionId, query);

  return (
    <div className="flex h-[calc(100vh-4rem)] gap-4 p-4">
      {/* Left panel: Timeline + Summary */}
      <div className="w-1/2 flex flex-col gap-4 overflow-hidden">
        <div className="flex items-center gap-3 shrink-0">
          <h2 className="text-lg font-semibold truncate">{state.query}</h2>
          <StatusBadge status={state.status} />
        </div>

        {state.error && (
          <div className="rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-2 text-sm text-destructive">
            {state.error}
          </div>
        )}

        {/* Prior knowledge */}
        {state.priorFacts.length > 0 && (
          <div className="rounded-lg border border-indigo-500/30 bg-indigo-500/5 p-3 shrink-0">
            <h3 className="text-sm font-medium text-indigo-400 mb-2">Prior Knowledge</h3>
            <ul className="space-y-1 text-sm text-muted-foreground">
              {state.priorFacts.slice(0, 5).map((fact, i) => (
                <li key={i}>{fact.fact}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Streaming summary */}
        <div className="flex-1 overflow-y-auto min-h-0">
          <StreamingSummary
            content={state.summary}
            isStreaming={state.status === "running"}
          />
        </div>

        {/* Timeline */}
        <div className="h-48 shrink-0 overflow-y-auto border-t border-border pt-3">
          <ResearchTimeline events={state.timeline} />
        </div>
      </div>

      {/* Right panel: Graph + Entities */}
      <div className="w-1/2 flex flex-col gap-4 overflow-hidden">
        <div className="flex-1 min-h-0 rounded-lg border border-border overflow-hidden">
          <SessionSubgraph
            entities={state.entities}
            relations={state.relations}
            graphData={state.graphData}
          />
        </div>

        {/* Discovered entities */}
        {state.entities.length > 0 && (
          <div className="h-48 shrink-0 overflow-y-auto">
            <h3 className="text-sm font-medium text-muted-foreground mb-2">
              Entities ({state.entities.length})
            </h3>
            <div className="grid grid-cols-2 gap-2">
              {state.entities.slice(0, 10).map((entity) => (
                <EntityCard key={entity.uuid} entity={entity} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    idle: "bg-muted text-muted-foreground",
    connecting: "bg-yellow-500/20 text-yellow-400",
    running: "bg-indigo-500/20 text-indigo-400",
    complete: "bg-green-500/20 text-green-400",
    error: "bg-destructive/20 text-destructive",
  };

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${colors[status] || colors.idle}`}
    >
      {status === "running" && (
        <span className="h-1.5 w-1.5 rounded-full bg-current animate-pulse" />
      )}
      {status}
    </span>
  );
}
