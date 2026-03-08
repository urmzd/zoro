"use client";

import { IconLoader2, IconSearch } from "@tabler/icons-react";
import { useState } from "react";

interface SearchSectionProps {
  onSearch: (query: string) => void;
  isStreaming: boolean;
}

export function SearchSection({ onSearch, isStreaming }: SearchSectionProps) {
  const [query, setQuery] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim() || isStreaming) return;
    onSearch(query.trim());
  }

  return (
    <section className="border-b border-border/50 bg-background/60 backdrop-blur-sm px-4 py-5">
      <div className="mx-auto max-w-4xl space-y-3">
        <h2 className="text-lg font-semibold tracking-tight text-foreground">Explore Knowledge</h2>
        <form onSubmit={handleSubmit} className="flex items-center gap-3">
          <div className="flex flex-1 items-center gap-3 rounded-2xl border border-border bg-background/80 backdrop-blur-sm px-4 py-3">
            {isStreaming ? (
              <IconLoader2 className="h-5 w-5 text-muted-foreground shrink-0 animate-spin" />
            ) : (
              <IconSearch className="h-5 w-5 text-muted-foreground shrink-0" />
            )}
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search your knowledge graph..."
              disabled={isStreaming}
              className="flex-1 bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground text-base"
            />
          </div>
          <button
            type="submit"
            disabled={isStreaming || !query.trim()}
            className="shrink-0 rounded-full bg-indigo-600 px-5 py-2.5 text-sm font-medium text-white transition-opacity disabled:opacity-50 hover:opacity-90"
          >
            Search
          </button>
        </form>
      </div>
    </section>
  );
}
