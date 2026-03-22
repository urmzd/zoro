"use client";

import { IconRefresh } from "@tabler/icons-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { getLogs, getStatus, type ServiceStatus } from "@/app/lib/api";

export default function LogsPage() {
  const [logs, setLogs] = useState("");
  const [status, setStatus] = useState<ServiceStatus | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const scrollRef = useRef<HTMLPreElement>(null);

  const refresh = useCallback(async () => {
    const [logText, svcStatus] = await Promise.all([getLogs(300), getStatus()]);
    setLogs(logText);
    setStatus(svcStatus);
  }, []);

  useEffect(() => {
    refresh();
    if (!autoRefresh) return;
    const interval = setInterval(refresh, 3000);
    return () => clearInterval(interval);
  }, [refresh, autoRefresh]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs]);

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-4 border-b border-border/50">
        <h1 className="text-lg font-semibold">Logs</h1>
        <div className="flex items-center gap-3">
          {status && (
            <div className="flex items-center gap-2 text-xs">
              <StatusDot label="SurrealDB" ok={status.surrealdb} />
              <StatusDot label="SearXNG" ok={status.searxng} />
              <StatusDot label="Ollama" ok={status.ollama} />
            </div>
          )}
          <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded"
            />
            Auto-refresh
          </label>
          <button
            type="button"
            onClick={refresh}
            className="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/50"
            title="Refresh"
          >
            <IconRefresh className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Log output */}
      <pre
        ref={scrollRef}
        className="flex-1 overflow-auto p-4 text-xs font-mono leading-relaxed text-muted-foreground bg-black/20"
      >
        {logs || "No logs available."}
      </pre>
    </div>
  );
}

function StatusDot({ label, ok }: { label: string; ok: boolean }) {
  return (
    <span className="flex items-center gap-1">
      <span
        className={`inline-block h-2 w-2 rounded-full ${ok ? "bg-emerald-500" : "bg-red-500"}`}
      />
      <span className={ok ? "text-foreground" : "text-red-400"}>{label}</span>
    </span>
  );
}
