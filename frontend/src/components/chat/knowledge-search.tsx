"use client";

import { IconSearch } from "@tabler/icons-react";
import { useState } from "react";
import { searchKnowledge } from "@/app/lib/api";

interface KnowledgeFact {
  uuid?: string;
  name?: string;
  fact?: string;
  source_node?: { name?: string };
  target_node?: { name?: string };
}

export function KnowledgeSearch() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<KnowledgeFact[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim() || loading) return;

    setLoading(true);
    setSearched(true);
    try {
      const data = await searchKnowledge(query.trim());
      setResults(data?.facts ?? []);
    } catch {
      setResults([]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-full max-w-2xl mx-auto space-y-4">
      <form onSubmit={handleSubmit}>
        <div className="flex items-center gap-3 rounded-2xl border border-border bg-background/80 backdrop-blur-sm px-4 py-3">
          <IconSearch className="h-5 w-5 text-muted-foreground shrink-0" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search your knowledge graph..."
            disabled={loading}
            className="flex-1 bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground text-base"
          />
          <button
            type="submit"
            disabled={loading || !query.trim()}
            className="shrink-0 rounded-full bg-purple-600 px-4 py-1.5 text-sm font-medium text-white transition-opacity disabled:opacity-50 hover:opacity-90"
          >
            {loading ? "Searching..." : "Search"}
          </button>
        </div>
      </form>

      {searched && !loading && (
        <div className="space-y-2">
          {results.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground py-4">
              No results found. Try a different query or build knowledge through chat first.
            </p>
          ) : (
            <div className="space-y-2 max-h-[40vh] overflow-y-auto">
              {results.map((fact, i) => (
                <div
                  key={fact.uuid ?? i}
                  className="rounded-lg border border-border/50 bg-card/50 px-4 py-3 text-sm"
                >
                  {(fact.source_node?.name || fact.target_node?.name) && (
                    <div className="flex items-center gap-2 mb-1">
                      {fact.source_node?.name && (
                        <span className="rounded-full bg-indigo-600/20 px-2 py-0.5 text-xs font-medium text-indigo-400">
                          {fact.source_node.name}
                        </span>
                      )}
                      {fact.source_node?.name && fact.target_node?.name && (
                        <span className="text-muted-foreground text-xs">&rarr;</span>
                      )}
                      {fact.target_node?.name && (
                        <span className="rounded-full bg-purple-600/20 px-2 py-0.5 text-xs font-medium text-purple-400">
                          {fact.target_node.name}
                        </span>
                      )}
                    </div>
                  )}
                  <p className="text-foreground/80">{fact.fact}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
