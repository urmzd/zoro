"use client";

import { useMemo, useState } from "react";
import type { ToolCall } from "@/app/lib/types";
import { cn } from "@/lib/utils";

const TOOL_CONFIG: Record<string, { icon: string; color: string; label: string }> = {
  web_search: {
    icon: "language",
    color: "text-[#ba9eff]",
    label: "WEB_SEARCH",
  },
  search_knowledge: {
    icon: "database",
    color: "text-[#699cff]",
    label: "SEARCH_KNOWLEDGE",
  },
  store_knowledge: {
    icon: "save",
    color: "text-emerald-400",
    label: "STORE_KNOWLEDGE",
  },
};

interface SearchResultItem {
  index: number;
  title: string;
  url: string;
  snippet: string;
}

function parseSearchResults(result: string): SearchResultItem[] | null {
  try {
    const parsed = JSON.parse(result);
    if (Array.isArray(parsed) && parsed.length > 0 && parsed[0].url) {
      return parsed;
    }
  } catch {
    // not JSON — fall through
  }
  return null;
}

function getDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return url;
  }
}

function WebSearchResults({ results }: { results: SearchResultItem[] }) {
  return (
    <div className="space-y-2">
      {results.map((r) => (
        <a
          key={r.index}
          href={r.url}
          target="_blank"
          rel="noopener noreferrer"
          className="flex gap-3 px-3 py-2.5 rounded-lg bg-white/[0.02] hover:bg-white/[0.06] border border-[#40485d]/10 hover:border-[#ba9eff]/20 transition-all group"
        >
          <span className="shrink-0 w-5 h-5 rounded-md bg-[#ba9eff]/10 text-[#ba9eff] text-[10px] font-bold flex items-center justify-center mt-0.5">
            {r.index}
          </span>
          <div className="min-w-0 flex-1">
            <div className="text-xs font-medium text-[#dee5ff] group-hover:text-[#ba9eff] transition-colors truncate">
              {r.title}
            </div>
            <div className="text-[10px] text-[#a3aac4]/60 truncate mt-0.5">{getDomain(r.url)}</div>
            <div className="text-[11px] text-[#a3aac4]/70 line-clamp-2 mt-1 leading-relaxed">
              {r.snippet}
            </div>
          </div>
          <span className="material-symbols-outlined text-sm text-[#a3aac4]/30 group-hover:text-[#ba9eff]/50 shrink-0 mt-0.5 transition-colors">
            open_in_new
          </span>
        </a>
      ))}
    </div>
  );
}

interface ToolCallCardProps {
  toolCall: ToolCall;
}

export function ToolCallCard({ toolCall }: ToolCallCardProps) {
  const [expanded, setExpanded] = useState(false);
  const config = TOOL_CONFIG[toolCall.name] || TOOL_CONFIG.web_search;
  const isLoading = !toolCall.result;

  const searchResults = useMemo(
    () =>
      toolCall.name === "web_search" && toolCall.result
        ? parseSearchResults(toolCall.result)
        : null,
    [toolCall.name, toolCall.result],
  );

  const hasRichResults = searchResults !== null;
  const resultCount = searchResults?.length ?? 0;

  return (
    <div className="glass-card rounded-xl overflow-hidden max-w-2xl">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        aria-expanded={expanded}
        className="flex items-center justify-between w-full px-4 py-3 bg-white/5"
      >
        <div className="flex items-center gap-3">
          <span className={cn("material-symbols-outlined text-lg", config.color)}>
            {config.icon}
          </span>
          <span className="text-xs font-headline font-bold uppercase tracking-wider text-[#dee5ff]">
            {config.label}
          </span>
          {hasRichResults && (
            <span className="text-[10px] text-[#a3aac4]/60">
              {resultCount} source{resultCount !== 1 ? "s" : ""}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {isLoading ? (
            <span className="text-[10px] text-[#a3aac4] italic">Running...</span>
          ) : (
            <span className="text-[10px] text-[#a3aac4]">Complete</span>
          )}
          <span className="material-symbols-outlined text-sm text-[#a3aac4] hover:text-[#ba9eff]">
            {expanded ? "expand_less" : "expand_more"}
          </span>
        </div>
      </button>
      {expanded && (
        <div className="px-4 py-3 border-t border-[#40485d]/5">
          {hasRichResults ? (
            <WebSearchResults results={searchResults} />
          ) : (
            <div className="text-xs font-mono text-[#a3aac4]/80">
              <div className="mb-2">
                <span className="text-[#699cff]">args:</span> {toolCall.arguments}
              </div>
              {toolCall.result && (
                <div className="max-h-40 overflow-y-auto">
                  <span className="text-[#ff716a]">result:</span> {toolCall.result}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
