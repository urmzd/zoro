"use client";

import { IconMessageCircle, IconSearch } from "@tabler/icons-react";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import type { ChatSessionSummary } from "@/app/lib/api";
import { listChatSessions, searchSessions } from "@/app/lib/api";
import { cn } from "@/lib/utils";

export function SidebarHistory() {
  const [sessions, setSessions] = useState<ChatSessionSummary[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const hasFetched = useRef(false);
  const router = useRouter();
  const searchParams = useSearchParams();
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const activeId = searchParams.get("id");

  const fetchSessions = useCallback(async (query = "") => {
    setLoading(true);
    try {
      const data = query ? await searchSessions(query) : await listChatSessions();
      setSessions(data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!hasFetched.current) {
      hasFetched.current = true;
      fetchSessions();
    }
  }, [fetchSessions]);

  // Refresh periodically when visible
  useEffect(() => {
    const interval = setInterval(() => fetchSessions(searchQuery), 10000);
    return () => clearInterval(interval);
  }, [fetchSessions, searchQuery]);

  function handleSearch(value: string) {
    setSearchQuery(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      fetchSessions(value);
    }, 300);
  }

  function handleSelect(id: string) {
    router.push(`/chat?id=${id}`);
  }

  return (
    <div className="px-2 py-2">
      <div className="flex items-center gap-1.5 px-2 mb-2">
        <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider flex-1">
          History
        </h3>
      </div>

      {/* Search input */}
      <div className="relative mb-2 px-1">
        <IconSearch className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => handleSearch(e.target.value)}
          placeholder="Search sessions..."
          className="w-full rounded-md border border-border/50 bg-background/50 pl-8 pr-3 py-1.5 text-xs outline-none placeholder:text-muted-foreground/50 focus:border-indigo-500/50"
        />
      </div>

      {/* Session list */}
      <div className="space-y-0.5">
        {loading && sessions.length === 0 ? (
          <p className="px-2 py-4 text-xs text-muted-foreground text-center">Loading...</p>
        ) : sessions.length === 0 ? (
          <p className="px-2 py-4 text-xs text-muted-foreground text-center">
            {searchQuery ? "No matches" : "No sessions yet"}
          </p>
        ) : (
          sessions.map((s) => (
            <button
              key={s.id}
              type="button"
              onClick={() => handleSelect(s.id ?? "")}
              className={cn(
                "w-full text-left rounded-md px-2 py-2 transition-colors hover:bg-muted/50",
                activeId === s.id && "bg-muted/70",
              )}
            >
              <div className="flex items-start gap-2">
                <IconMessageCircle className="h-3.5 w-3.5 mt-0.5 shrink-0 text-muted-foreground" />
                <div className="min-w-0 flex-1">
                  <p className="text-xs text-foreground truncate">{s.preview || "Empty session"}</p>
                  <p className="text-[10px] text-muted-foreground mt-0.5">
                    {s.message_count} msgs
                    {" \u00b7 "}
                    {formatRelativeTime(s.created_at ?? "")}
                  </p>
                </div>
              </div>
            </button>
          ))
        )}
      </div>
    </div>
  );
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);

  if (diffMin < 1) return "just now";
  if (diffMin < 60) return `${diffMin}m ago`;

  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;

  const diffDay = Math.floor(diffHr / 24);
  if (diffDay < 7) return `${diffDay}d ago`;

  return date.toLocaleDateString();
}
