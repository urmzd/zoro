"use client";

import { IconSearch } from "@tabler/icons-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { HoverBorderGradient } from "@/components/ui/hover-border-gradient";

export function SearchForm() {
  const [query, setQuery] = useState("");
  const router = useRouter();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim()) return;

    // Navigate to research page — the SSE stream starts in useResearchStream
    const id = crypto.randomUUID();
    router.push(`/research?id=${id}&q=${encodeURIComponent(query.trim())}`);
  }

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-2xl mx-auto">
      <HoverBorderGradient containerClassName="w-full" className="w-full px-4 py-2 gap-3">
        <IconSearch className="h-5 w-5 text-muted-foreground shrink-0" />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="What would you like to research?"
          className="flex-1 bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground text-base"
        />
        <button
          type="submit"
          disabled={!query.trim()}
          className="shrink-0 rounded-full bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground transition-opacity disabled:opacity-50 hover:opacity-90"
        >
          Research
        </button>
      </HoverBorderGradient>
    </form>
  );
}
