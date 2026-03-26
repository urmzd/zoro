"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  classifyIntent,
  createChatSession,
  getAutocompleteSuggestions,
  searchKnowledge,
} from "@/app/lib/api";
import type { Fact } from "@/app/lib/types";
import { KnowledgeResults } from "@/components/chat/knowledge-results";
import { GraphErrorBoundary } from "@/components/graph/graph-error-boundary";
import { useChatStore } from "@/lib/stores/chat-store";

type Status =
  | "idle"
  | "autocompleting"
  | "classifying"
  | "starting_chat"
  | "searching_knowledge"
  | "done";

const STATUS_LABELS: Partial<Record<Status, string>> = {
  classifying: "Understanding your query...",
  starting_chat: "Starting session...",
  searching_knowledge: "Searching knowledge...",
};

export default function Home() {
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState<Status>("idle");
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [ghostText, setGhostText] = useState("");
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [knowledgeResults, setKnowledgeResults] = useState<Fact[]>([]);
  const [knowledgeSearched, setKnowledgeSearched] = useState(false);

  const router = useRouter();
  const abortRef = useRef<AbortController | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const inputRef = useRef<HTMLInputElement>(null);
  const suggestionsRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        suggestionsRef.current &&
        !suggestionsRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setShowSuggestions(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
      if (abortRef.current) abortRef.current.abort();
    };
  }, []);

  const fetchSuggestions = useCallback((value: string) => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    if (abortRef.current) abortRef.current.abort();

    if (value.trim().length < 2) {
      setSuggestions([]);
      setGhostText("");
      setShowSuggestions(false);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      const controller = new AbortController();
      abortRef.current = controller;
      try {
        const results = await getAutocompleteSuggestions(value.trim(), controller.signal);
        if (!controller.signal.aborted) {
          setSuggestions(results);
          setShowSuggestions(results.length > 0);
          const currentQuery = inputRef.current?.value ?? "";
          const match = results.find((s) => s.toLowerCase().startsWith(currentQuery.toLowerCase()));
          setGhostText(match ?? "");
        }
      } catch {
        // aborted or network error
      }
    }, 300);
  }, []);

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    const value = e.target.value;
    setQuery(value);
    setKnowledgeSearched(false);
    setKnowledgeResults([]);
    fetchSuggestions(value);
  }

  function selectSuggestion(suggestion: string) {
    setQuery(suggestion);
    setGhostText("");
    setShowSuggestions(false);
    setSuggestions([]);
    handleSubmit(undefined, suggestion);
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Tab" && ghostText && ghostText !== query) {
      e.preventDefault();
      setQuery(ghostText);
      setGhostText("");
      setShowSuggestions(false);
    }
  }

  async function handleSubmit(e?: React.FormEvent, overrideQuery?: string) {
    e?.preventDefault();
    const q = (overrideQuery ?? query).trim();
    if (!q || status !== "idle") return;

    setShowSuggestions(false);
    setSuggestions([]);

    setStatus("classifying");
    let action: "chat" | "knowledge_search";
    try {
      const intent = await classifyIntent(q);
      action = intent.action;
    } catch {
      action = "chat";
    }

    if (action === "chat") {
      setStatus("starting_chat");
      try {
        const { id } = await createChatSession();
        useChatStore.getState().reset();
        useChatStore.getState().setPendingQuery(q);
        router.push(`/chat?id=${id}`);
      } catch {
        setStatus("idle");
      }
    } else {
      setStatus("searching_knowledge");
      try {
        const data = await searchKnowledge(q);
        setKnowledgeResults(data?.facts ?? []);
        setKnowledgeSearched(true);
      } catch {
        setKnowledgeResults([]);
        setKnowledgeSearched(true);
      }
      setStatus("idle");
    }
  }

  const isLoading = status !== "idle";
  const statusLabel = STATUS_LABELS[status];

  return (
    <section className="flex-1 flex flex-col items-center justify-center px-6 py-24 relative z-10">
      {/* Background Luminous Core */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-[#ba9eff]/5 rounded-full blur-[150px] pointer-events-none" />

      <div className="w-full max-w-4xl text-center mb-16">
        <h2 className="font-headline text-5xl md:text-7xl font-bold tracking-tight mb-6">
          What should we{" "}
          <span className="zoro-gradient-text">uncover</span> today?
        </h2>
        <p className="text-[#a3aac4] text-lg max-w-2xl mx-auto font-light">
          Query the Luminous Intelligence core for structured facts or creative dialogue.
        </p>
      </div>

      {/* Search Bar */}
      <div className="w-full max-w-3xl relative">
        <form onSubmit={handleSubmit}>
          <div className="glass-panel p-1.5 rounded-2xl luminous-glow">
            <div className="flex items-center bg-black rounded-[14px] px-6 py-4">
              <span className="material-symbols-outlined text-[#ba9eff] text-2xl mr-4">
                {isLoading ? "progress_activity" : "search"}
              </span>
              <div className="relative flex-1">
                {ghostText &&
                  query &&
                  ghostText.toLowerCase().startsWith(query.toLowerCase()) &&
                  ghostText !== query && (
                    <span
                      aria-hidden
                      className="pointer-events-none absolute inset-0 flex items-center text-xl text-[#a3aac4]/20 select-none whitespace-nowrap overflow-hidden font-light"
                    >
                      <span className="invisible">{query}</span>
                      <span>{ghostText.slice(query.length)}</span>
                    </span>
                  )}
                <input
                  ref={inputRef}
                  type="text"
                  value={query}
                  onChange={handleInputChange}
                  onKeyDown={handleKeyDown}
                  onFocus={() => {
                    if (suggestions.length > 0) setShowSuggestions(true);
                  }}
                  placeholder="Search the knowledge graph..."
                  disabled={isLoading}
                  autoComplete="off"
                  className="relative w-full bg-transparent border-none outline-none text-xl text-[#dee5ff] placeholder:text-[#a3aac4]/30 font-light focus:ring-0"
                />
              </div>
              <div className="flex items-center gap-2">
                <span className="bg-[#141f38] text-[10px] text-[#a3aac4] px-2 py-1 rounded font-mono border border-[#40485d]/20">
                  ⌘ K
                </span>
              </div>
            </div>
          </div>
        </form>

        {/* Intent indicator */}
        {status === "classifying" && (
          <div className="absolute -bottom-10 left-6 flex items-center gap-3">
            <div className="flex items-center gap-2 px-3 py-1 bg-[#0f1930] rounded-full text-[10px] font-bold uppercase tracking-wider text-[#ba9eff] border border-[#ba9eff]/20">
              <span className="w-1.5 h-1.5 rounded-full bg-[#ba9eff] animate-pulse" />
              Classifying intent...
            </div>
          </div>
        )}

        {/* Autocomplete dropdown */}
        {showSuggestions && suggestions.length > 0 && (
          <div
            ref={suggestionsRef}
            className="absolute top-full left-0 right-0 mt-2 rounded-xl glass-card shadow-lg overflow-hidden z-20"
          >
            {suggestions.map((s) => (
              <button
                key={s}
                type="button"
                onClick={() => selectSuggestion(s)}
                className="w-full text-left px-6 py-3 text-sm text-[#dee5ff]/80 hover:bg-[#1f2b49] transition-colors"
              >
                {s}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Loading status */}
      {statusLabel && (
        <p className="text-sm text-[#a3aac4] animate-pulse mt-8">{statusLabel}</p>
      )}

      {/* Knowledge results */}
      {knowledgeSearched && (
        <div className="w-full max-w-4xl mt-16">
          <GraphErrorBoundary>
            <KnowledgeResults
              results={knowledgeResults}
              searched={knowledgeSearched}
              loading={status === "searching_knowledge"}
            />
          </GraphErrorBoundary>
        </div>
      )}
    </section>
  );
}
