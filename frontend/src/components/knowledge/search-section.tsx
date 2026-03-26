"use client";

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
    <section className="py-6 shrink-0 px-8">
      <div className="max-w-4xl mx-auto relative group">
        <div className="absolute -inset-1 bg-gradient-to-r from-[#ba9eff]/20 via-[#699cff]/10 to-[#ff716a]/20 rounded-2xl blur-xl opacity-50 group-focus-within:opacity-100 transition-opacity" />
        <form onSubmit={handleSubmit}>
          <div className="relative flex items-center bg-black border border-[#40485d]/30 rounded-2xl px-6 py-4 luminous-glow">
            <span className="material-symbols-outlined text-[#8455ef] mr-4">psychology</span>
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Explore neural pathways... e.g., 'How are the Q3 market trends linked to sustainable logistics?'"
              disabled={isStreaming}
              className="flex-1 bg-transparent border-none focus:ring-0 text-[#dee5ff] placeholder-[#a3aac4]/50 text-lg font-light"
            />
            <div className="flex items-center gap-3">
              <span className="px-2 py-1 rounded bg-[#141f38] text-[10px] text-[#a3aac4] font-bold border border-[#40485d]/20">
                ⌘ K
              </span>
              <button
                type="submit"
                disabled={isStreaming || !query.trim()}
                className="p-2 bg-[#ba9eff] rounded-lg text-black hover:bg-[#a27cff] transition-colors disabled:opacity-50"
              >
                <span className="material-symbols-outlined">arrow_forward</span>
              </button>
            </div>
          </div>
        </form>
      </div>
    </section>
  );
}
