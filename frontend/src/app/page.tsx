"use client";

import { IconLoader2, IconSearch } from "@tabler/icons-react";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  classifyIntent,
  createChatSession,
  getAutocompleteSuggestions,
  searchKnowledge,
} from "@/app/lib/api";
import { KnowledgeResults } from "@/components/chat/knowledge-results";
import { BackgroundBeams } from "@/components/ui/background-beams";
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
  const [knowledgeResults, setKnowledgeResults] = useState<any[]>([]);
  const [knowledgeSearched, setKnowledgeSearched] = useState(false);

  const router = useRouter();
  const abortRef = useRef<AbortController | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const inputRef = useRef<HTMLInputElement>(null);
  const suggestionsRef = useRef<HTMLDivElement>(null);

  // Close suggestions when clicking outside
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
          // Pick the best ghost text: first suggestion that extends the current input
          const currentQuery = inputRef.current?.value ?? "";
          const match = results.find((s) => s.toLowerCase().startsWith(currentQuery.toLowerCase()));
          setGhostText(match ?? "");
        }
      } catch {
        // aborted or network error — ignore
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

    // Step 1: Classify intent
    setStatus("classifying");
    let action: "chat" | "knowledge_search";
    try {
      const intent = await classifyIntent(q);
      action = intent.action;
    } catch {
      action = "chat";
    }

    // Step 2: Route based on intent
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
    <main className="relative flex min-h-screen flex-col items-center justify-center px-4">
      <BackgroundBeams />
      <div className="relative z-10 flex flex-col items-center gap-8 w-full max-w-3xl">
        <div className="text-center space-y-3">
          <h1 className="text-6xl font-bold tracking-tight bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
            Zoro
          </h1>
          <p className="text-muted-foreground text-lg">Find The Truth.</p>
        </div>

        {/* Unified input */}
        <div className="w-full max-w-2xl mx-auto relative">
          <form onSubmit={handleSubmit}>
            <div className="flex items-center gap-3 rounded-2xl border border-border bg-background/80 backdrop-blur-sm px-4 py-3">
              {isLoading ? (
                <IconLoader2 className="h-5 w-5 text-muted-foreground shrink-0 animate-spin" />
              ) : (
                <IconSearch className="h-5 w-5 text-muted-foreground shrink-0" />
              )}
              <div className="relative flex-1">
                {/* Ghost text overlay */}
                {ghostText &&
                  query &&
                  ghostText.toLowerCase().startsWith(query.toLowerCase()) &&
                  ghostText !== query && (
                    <span
                      aria-hidden
                      className="pointer-events-none absolute inset-0 flex items-center text-base text-muted-foreground/25 select-none whitespace-nowrap overflow-hidden"
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
                  placeholder="Ask anything — Zoro will figure out the rest"
                  disabled={isLoading}
                  autoComplete="off"
                  className="relative w-full bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground text-base"
                />
              </div>
              <button
                type="submit"
                disabled={isLoading || !query.trim()}
                className="shrink-0 rounded-full bg-indigo-600 px-4 py-1.5 text-sm font-medium text-white transition-opacity disabled:opacity-50 hover:opacity-90"
              >
                Go
              </button>
            </div>
          </form>

          {/* Autocomplete dropdown */}
          {showSuggestions && suggestions.length > 0 && (
            <div
              ref={suggestionsRef}
              className="absolute top-full left-0 right-0 mt-1 rounded-xl border border-border bg-background/95 backdrop-blur-sm shadow-lg overflow-hidden z-20"
            >
              {suggestions.map((s) => (
                <button
                  key={s}
                  type="button"
                  onClick={() => selectSuggestion(s)}
                  className="w-full text-left px-4 py-2.5 text-sm text-foreground/80 hover:bg-muted/50 transition-colors"
                >
                  {s}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Loading status */}
        {statusLabel && (
          <p className="text-sm text-muted-foreground animate-pulse">{statusLabel}</p>
        )}

        {/* Inline knowledge results */}
        {knowledgeSearched && (
          <div className="w-full max-w-2xl mx-auto">
            <KnowledgeResults
              results={knowledgeResults}
              searched={knowledgeSearched}
              loading={status === "searching_knowledge"}
            />
          </div>
        )}
      </div>
    </main>
  );
}
