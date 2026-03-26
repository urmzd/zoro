"use client";

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

  useEffect(() => {
    const interval = setInterval(() => fetchSessions(searchQuery), 10000);
    return () => {
      clearInterval(interval);
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
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
      <p className="text-[11px] font-bold text-[#6d758c] uppercase tracking-wider mb-2 px-2">
        Recent Sessions
      </p>

      {/* Search */}
      <div className="relative mb-3 px-1">
        <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[#a3aac4] text-sm">
          search
        </span>
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => handleSearch(e.target.value)}
          placeholder="Search sessions..."
          className="w-full rounded-md bg-[#060e20] border-none pl-8 pr-3 py-1.5 text-xs text-[#dee5ff] outline-none placeholder:text-[#a3aac4]/40 focus:ring-1 focus:ring-[#8455ef]/40"
        />
      </div>

      {/* Sessions */}
      <div className="space-y-1">
        {loading && sessions.length === 0 ? (
          <p className="px-2 py-4 text-xs text-[#a3aac4] text-center">Loading...</p>
        ) : sessions.length === 0 ? (
          <p className="px-2 py-4 text-xs text-[#a3aac4] text-center">
            {searchQuery ? "No matches" : "No sessions yet"}
          </p>
        ) : (
          sessions
            .filter((s) => s.id)
            .map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => handleSelect(s.id)}
                className={cn(
                  "group w-full text-left px-2.5 py-1.5 rounded-md transition-all",
                  activeId === s.id ? "bg-[#141f38]" : "hover:bg-[#141f38]",
                )}
              >
                <p className="text-[13px] text-[#a3aac4] group-hover:text-[#dee5ff] truncate">
                  {s.preview || "Empty session"}
                </p>
                <p className="text-[11px] text-[#6d758c] mt-0.5">
                  {s.message_count} msgs · {formatRelativeTime(s.created_at ?? "")}
                </p>
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
