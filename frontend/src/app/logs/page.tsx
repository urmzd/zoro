"use client";

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
  }, []);

  return (
    <div className="flex-1 overflow-y-auto px-8 pb-12 pt-6">
      {/* Title */}
      <div className="mb-10 flex justify-between items-end">
        <div>
          <h2 className="text-4xl font-headline font-bold text-white tracking-tight mb-2">
            System Logs
          </h2>
          <p className="text-[#a3aac4]">Real-time developer telemetry and service diagnostics.</p>
        </div>
        <div className="flex items-center gap-4 bg-[#050a18] p-2 rounded-2xl border border-[#40485d]/10">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-xl cursor-pointer group">
            <div className="relative flex h-2 w-2">
              {autoRefresh && (
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#ba9eff] opacity-75" />
              )}
              <span className={`relative inline-flex rounded-full h-2 w-2 ${autoRefresh ? "bg-[#ba9eff]" : "bg-[#6d758c]"}`} />
            </div>
            <label className="text-sm font-medium cursor-pointer flex items-center gap-2">
              Auto-refresh
              <button
                type="button"
                onClick={() => setAutoRefresh(!autoRefresh)}
                className={`w-8 h-4 rounded-full relative transition-colors ${autoRefresh ? "bg-[#ba9eff]" : "bg-[#40485d]"}`}
              >
                <div className={`absolute top-0.5 w-3 h-3 bg-white rounded-full transition-all ${autoRefresh ? "right-0.5" : "left-0.5"}`} />
              </button>
            </label>
          </div>
          <button
            type="button"
            onClick={refresh}
            className="flex items-center gap-2 px-4 py-1.5 bg-[#1f2b49] rounded-xl hover:scale-[1.02] active:scale-95 transition-all text-sm font-bold"
          >
            <span className="material-symbols-outlined text-sm">refresh</span>
            Refresh
          </button>
        </div>
      </div>

      {/* Bento Grid */}
      <div className="grid grid-cols-12 gap-6 mb-8">
        {/* Service status cards */}
        <div className="col-span-12 lg:col-span-4 grid grid-cols-1 gap-4">
          {status && (
            <>
              <StatusCard name="SurrealDB" description="Vector Engine" icon="database" iconColor="text-[#ba9eff]" ok={status.surrealdb} />
              <StatusCard name="SearXNG" description="Metasearch Node" icon="travel_explore" iconColor="text-[#699cff]" ok={status.searxng} />
              <StatusCard name="Ollama" description="Inference Backend" icon="memory" iconColor="text-[#ff716a]" ok={status.ollama} />
            </>
          )}
        </div>

        {/* System throughput placeholder */}
        <div className="col-span-12 lg:col-span-8 glass-panel p-8 rounded-2xl border border-[#40485d]/10 relative overflow-hidden">
          <div className="absolute top-0 right-0 p-8 opacity-10">
            <span className="material-symbols-outlined text-[120px]">analytics</span>
          </div>
          <h4 className="font-headline font-bold text-xl mb-6 flex items-center gap-2">
            <span className="material-symbols-outlined text-[#ba9eff]">monitoring</span>
            System Throughput
          </h4>
          <div className="flex items-end gap-1 h-32 mb-6">
            {[40, 55, 30, 80, 65, 45, 90, 35, 20, 55, 75, 60, 85, 40].map((h, i) => (
              <div
                key={i}
                className="flex-1 bg-[#ba9eff]/20 rounded-t-md hover:bg-[#ba9eff]/40 transition-all"
                style={{ height: `${h}%` }}
              />
            ))}
          </div>
        </div>
      </div>

      {/* Log console */}
      <div className="glass-panel rounded-3xl border border-[#40485d]/10 overflow-hidden flex flex-col shadow-2xl">
        <div className="bg-[#141f38] px-6 py-4 border-b border-[#40485d]/10 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex gap-1.5">
              <div className="w-3 h-3 rounded-full bg-[#d73357]/40" />
              <div className="w-3 h-3 rounded-full bg-[#ff716a]/40" />
              <div className="w-3 h-3 rounded-full bg-[#ba9eff]/40" />
            </div>
            <span className="text-xs font-mono text-[#a3aac4] ml-4 tracking-tight">
              zoro-main-instance.log
            </span>
          </div>
        </div>
        <pre
          ref={scrollRef}
          className="bg-black p-6 h-[500px] overflow-y-auto font-mono text-sm leading-relaxed text-[#dee5ff]/90"
        >
          {logs || "No logs available. Waiting for data..."}
        </pre>
      </div>
    </div>
  );
}

function StatusCard({
  name,
  description,
  icon,
  iconColor,
  ok,
}: {
  name: string;
  description: string;
  icon: string;
  iconColor: string;
  ok: boolean;
}) {
  return (
    <div className="glass-panel p-6 rounded-2xl border border-[#40485d]/10 flex items-center justify-between">
      <div className="flex items-center gap-4">
        <div className="w-12 h-12 rounded-xl bg-[#0f1930] flex items-center justify-center">
          <span
            className={`material-symbols-outlined ${iconColor}`}
            style={{ fontVariationSettings: "'FILL' 1" }}
          >
            {icon}
          </span>
        </div>
        <div>
          <h4 className="font-bold text-white">{name}</h4>
          <p className="text-xs text-[#a3aac4]">{description}</p>
        </div>
      </div>
      <div className={`flex items-center gap-2 px-3 py-1 rounded-full ${ok ? "bg-emerald-500/10" : "bg-red-500/10"}`}>
        <div className={`h-1.5 w-1.5 rounded-full ${ok ? "bg-emerald-500" : "bg-red-500"}`} />
        <span className={`text-[10px] font-bold uppercase ${ok ? "text-emerald-500" : "text-red-500"}`}>
          {ok ? "Active" : "Down"}
        </span>
      </div>
    </div>
  );
}
