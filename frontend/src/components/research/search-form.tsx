"use client";

import { IconSearch } from "@tabler/icons-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { startResearch } from "@/app/lib/api";
import { HoverBorderGradient } from "@/components/ui/hover-border-gradient";

export function SearchForm() {
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim() || loading) return;

    setLoading(true);
    try {
      const id = await startResearch(query.trim());
      router.push(`/research/${id}?q=${encodeURIComponent(query.trim())}`);
    } catch {
      setLoading(false);
    }
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
          disabled={loading}
          className="flex-1 bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground text-base"
        />
        <button
          type="submit"
          disabled={loading || !query.trim()}
          className="shrink-0 rounded-full bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground transition-opacity disabled:opacity-50 hover:opacity-90"
        >
          {loading ? "Starting..." : "Research"}
        </button>
      </HoverBorderGradient>
    </form>
  );
}
