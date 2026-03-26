"use client";

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
        <div className="flex items-center gap-3 rounded-2xl bg-black border border-[#40485d]/20 px-4 py-3">
          <span className="material-symbols-outlined text-[#ba9eff] shrink-0">search</span>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search your knowledge graph..."
            disabled={loading}
            className="flex-1 bg-transparent border-none outline-none text-[#dee5ff] placeholder:text-[#a3aac4]/40 text-base focus:ring-0"
          />
          <button
            type="submit"
            disabled={loading || !query.trim()}
            className="shrink-0 rounded-full zoro-gradient-bg px-4 py-1.5 text-sm font-medium text-black transition-opacity disabled:opacity-50"
          >
            {loading ? "Searching..." : "Search"}
          </button>
        </div>
      </form>

      {searched && !loading && (
        <div className="space-y-2">
          {results.length === 0 ? (
            <p className="text-center text-sm text-[#a3aac4] py-4">
              No results found. Try a different query or build knowledge through chat first.
            </p>
          ) : (
            <div className="space-y-2 max-h-[40vh] overflow-y-auto">
              {results.map((fact, i) => (
                <div key={fact.uuid ?? i} className="rounded-xl glass-card px-4 py-3 text-sm">
                  {(fact.source_node?.name || fact.target_node?.name) && (
                    <div className="flex items-center gap-2 mb-1">
                      {fact.source_node?.name && (
                        <span className="rounded-full bg-[#ba9eff]/20 px-2 py-0.5 text-xs font-medium text-[#ba9eff]">
                          {fact.source_node.name}
                        </span>
                      )}
                      {fact.source_node?.name && fact.target_node?.name && (
                        <span className="text-[#a3aac4] text-xs">&rarr;</span>
                      )}
                      {fact.target_node?.name && (
                        <span className="rounded-full bg-[#699cff]/20 px-2 py-0.5 text-xs font-medium text-[#699cff]">
                          {fact.target_node.name}
                        </span>
                      )}
                    </div>
                  )}
                  <p className="text-[#dee5ff]/80">{fact.fact}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
