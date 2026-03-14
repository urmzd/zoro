"use client";

import {
  IconChevronLeft,
  IconChevronRight,
  IconGraph,
  IconMessageCirclePlus,
} from "@tabler/icons-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Suspense, useCallback, useEffect, useState } from "react";
import { createChatSession } from "@/app/lib/api";
import { cn } from "@/lib/utils";
import { SidebarHistory } from "./sidebar-history";
import { SidebarKnowledge } from "./sidebar-knowledge";
import { SidebarSettings } from "./sidebar-settings";

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(() => {
    if (typeof window === "undefined") return false;
    return localStorage.getItem("sidebar-collapsed") === "true";
  });
  const router = useRouter();

  useEffect(() => {
    localStorage.setItem("sidebar-collapsed", String(collapsed));
  }, [collapsed]);

  const handleNewChat = useCallback(async () => {
    try {
      const { id } = await createChatSession();
      router.push(`/chat?id=${id}`);
    } catch {
      // ignore
    }
  }, [router]);

  return (
    <aside
      className={cn(
        "flex flex-col border-r border-border bg-background/50 backdrop-blur-sm transition-all duration-200",
        collapsed ? "w-12" : "w-64",
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-3 border-b border-border/50">
        {!collapsed && (
          <Link
            href="/"
            className="text-lg font-bold tracking-tight bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent"
          >
            Zoro
          </Link>
        )}
        <button
          type="button"
          onClick={() => setCollapsed(!collapsed)}
          className="text-muted-foreground hover:text-foreground p-1 rounded-md hover:bg-muted/50"
        >
          {collapsed ? (
            <IconChevronRight className="h-4 w-4" />
          ) : (
            <IconChevronLeft className="h-4 w-4" />
          )}
        </button>
      </div>

      {/* New Chat button */}
      <div className="px-2 py-2">
        <button
          type="button"
          onClick={handleNewChat}
          className={cn(
            "flex items-center gap-2 w-full rounded-lg px-3 py-2 text-sm font-medium transition-colors",
            "bg-indigo-600 text-white hover:bg-indigo-700",
            collapsed && "justify-center px-0",
          )}
        >
          <IconMessageCirclePlus className="h-4 w-4 shrink-0" />
          {!collapsed && "New Chat"}
        </button>
      </div>

      {/* Scrollable content */}
      <div className="flex-1 overflow-y-auto min-h-0">
        {!collapsed && (
          <Suspense
            fallback={<div className="px-4 py-2 text-xs text-muted-foreground">Loading...</div>}
          >
            <SidebarHistory />
          </Suspense>
        )}
      </div>

      {/* Bottom section */}
      <div className="mt-auto border-t border-border/50">
        {!collapsed && <SidebarKnowledge />}
        {!collapsed && <SidebarSettings />}
        {collapsed && (
          <div className="flex flex-col items-center gap-1 py-2">
            <Link
              href="/knowledge"
              className="p-2 text-muted-foreground hover:text-foreground rounded-md hover:bg-muted/50"
              title="Knowledge Graph"
            >
              <IconGraph className="h-4 w-4" />
            </Link>
          </div>
        )}
      </div>
    </aside>
  );
}
