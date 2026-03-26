"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

export function SearchForm() {
  const [query, setQuery] = useState("");
  const router = useRouter();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim()) return;

    const id = crypto.randomUUID();
    router.push(`/research?id=${id}&q=${encodeURIComponent(query.trim())}`);
  }

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-2xl mx-auto">
      <div className="glass-panel p-1.5 rounded-2xl luminous-glow">
        <div className="flex items-center bg-black rounded-[14px] px-6 py-4 gap-3">
          <span className="material-symbols-outlined text-[#ba9eff] text-xl">search</span>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="What would you like to research?"
            className="flex-1 bg-transparent border-none outline-none text-[#dee5ff] placeholder:text-[#a3aac4]/40 text-base focus:ring-0"
          />
          <button
            type="submit"
            disabled={!query.trim()}
            className="shrink-0 rounded-xl zoro-gradient-bg px-5 py-2 text-sm font-bold text-black transition-opacity disabled:opacity-50 hover:scale-105 active:scale-95 transition-all"
          >
            Research
          </button>
        </div>
      </div>
    </form>
  );
}
