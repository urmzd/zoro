"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { IconHistory, IconX, IconMessageCircle } from "@tabler/icons-react";
import { listChatSessions } from "@/app/lib/api";
import type { ChatSessionSummary } from "@/app/lib/types";
import { cn } from "@/lib/utils";

export function SessionsSidebar() {
  const [open, setOpen] = useState(false);
  const [sessions, setSessions] = useState<ChatSessionSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const hasFetched = useRef(false);
  const router = useRouter();
  const pathname = usePathname();

  const fetchSessions = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listChatSessions();
      setSessions(data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open && !hasFetched.current) {
      hasFetched.current = true;
      fetchSessions();
    }
  }, [open, fetchSessions]);

  // Refresh when sidebar is re-opened
  useEffect(() => {
    if (open && hasFetched.current) {
      fetchSessions();
    }
  }, [open, fetchSessions]);

  function handleSelect(id: string) {
    router.push(`/chat/${id}`);
    setOpen(false);
  }

  const activeId = pathname.startsWith("/chat/")
    ? pathname.split("/")[2]
    : null;

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          "flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
          open
            ? "bg-muted text-foreground"
            : "text-muted-foreground hover:text-foreground hover:bg-muted/50",
        )}
      >
        <IconHistory className="h-4 w-4" />
        Sessions
      </button>

      {open && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-40 bg-black/20 backdrop-blur-sm"
            onClick={() => setOpen(false)}
            onKeyDown={() => {}}
          />

          {/* Panel */}
          <div className="fixed top-12 right-0 z-50 h-[calc(100vh-3rem)] w-80 border-l border-border bg-background shadow-xl flex flex-col">
            <div className="flex items-center justify-between px-4 py-3 border-b border-border">
              <h2 className="text-sm font-medium">Chat Sessions</h2>
              <button
                type="button"
                onClick={() => setOpen(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <IconX className="h-4 w-4" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto">
              {loading && sessions.length === 0 ? (
                <div className="px-4 py-8 text-center text-sm text-muted-foreground">
                  Loading...
                </div>
              ) : sessions.length === 0 ? (
                <div className="px-4 py-8 text-center text-sm text-muted-foreground">
                  No sessions yet. Start a conversation!
                </div>
              ) : (
                <div className="py-1">
                  {sessions.map((s) => (
                    <button
                      key={s.id}
                      type="button"
                      onClick={() => handleSelect(s.id)}
                      className={cn(
                        "w-full text-left px-4 py-3 transition-colors hover:bg-muted/50 border-b border-border/30",
                        activeId === s.id && "bg-muted/70",
                      )}
                    >
                      <div className="flex items-start gap-2.5">
                        <IconMessageCircle className="h-4 w-4 mt-0.5 shrink-0 text-muted-foreground" />
                        <div className="min-w-0 flex-1">
                          <p className="text-sm text-foreground truncate">
                            {s.preview || "Empty session"}
                          </p>
                          <p className="text-xs text-muted-foreground mt-0.5">
                            {s.message_count} messages
                            {" \u00b7 "}
                            {formatRelativeTime(s.created_at)}
                          </p>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </>
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
