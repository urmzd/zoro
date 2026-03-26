"use client";

import type { ToolCall } from "@/app/lib/types";

interface SourceInfo {
  index: number;
  url: string;
  title: string;
}

function extractSources(toolCalls: ToolCall[]): Map<number, SourceInfo> {
  const sources = new Map<number, SourceInfo>();
  for (const tc of toolCalls) {
    if (tc.name !== "web_search" || !tc.result) continue;
    try {
      const results = JSON.parse(tc.result);
      if (!Array.isArray(results)) continue;
      for (const r of results) {
        if (r.index && r.url) {
          sources.set(r.index, { index: r.index, url: r.url, title: r.title || "" });
        }
      }
    } catch {
      // not JSON
    }
  }
  return sources;
}

const CITATION_RE = /\[(\d+)\]/g;

interface CitationTextProps {
  content: string;
  toolCalls: ToolCall[];
}

export function injectCitationLinks(content: string, toolCalls: ToolCall[]): string {
  const sources = extractSources(toolCalls);
  if (sources.size === 0) return content;

  return content.replace(CITATION_RE, (match, num) => {
    const idx = parseInt(num, 10);
    const source = sources.get(idx);
    if (!source) return match;
    return `[<sup>${idx}</sup>](${source.url} "${source.title}")`;
  });
}
