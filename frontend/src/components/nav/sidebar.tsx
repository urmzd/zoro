"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Suspense, useCallback, useEffect, useState } from "react";
import { createChatSession } from "@/app/lib/api";
import { cn } from "@/lib/utils";
import { SidebarHistory } from "./sidebar-history";

const NAV_ITEMS = [
  { href: "/knowledge", icon: "database", label: "Knowledge" },
  { href: "/chat", icon: "chat_bubble", label: "Chat", matchPrefix: true },
  { href: "/research", icon: "insights", label: "Insights" },
];

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const saved = localStorage.getItem("sidebar-collapsed");
    if (saved === "true") setCollapsed(true);
  }, []);

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
        "flex flex-col h-full py-4 shrink-0 bg-[#050a18] luminous-glow z-50 transition-all duration-300 ease-in-out",
        collapsed ? "w-12 px-1.5" : "w-56 px-3",
      )}
    >
      {/* Brand + Collapse toggle */}
      <div className={cn("mb-4", collapsed ? "px-0" : "px-1")}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3 shrink-0">
            {!collapsed && (
              <Link href="/">
                <h1 className="text-lg font-bold tracking-tight text-white font-headline">
                  Zoro
                </h1>
              </Link>
            )}
          </div>
          <button
            type="button"
            onClick={() => setCollapsed(!collapsed)}
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
            className={cn(
              "text-[#a3aac4]/60 hover:text-[#dee5ff] transition-colors duration-200 rounded-lg hover:bg-[#1f2b49]/40 p-1",
              collapsed && "mx-auto mt-2",
            )}
          >
            <span className="material-symbols-outlined text-[18px]">
              {collapsed ? "chevron_right" : "chevron_left"}
            </span>
          </button>
        </div>
      </div>

      {/* New Chat */}
      <button
        type="button"
        onClick={handleNewChat}
        className={cn(
          "mb-3 zoro-gradient-bg text-black rounded-md font-bold flex items-center justify-center gap-2 hover:scale-[1.02] transition-all duration-200 active:scale-95 shrink-0",
          collapsed ? "w-8 h-8 mx-auto p-0" : "w-full py-1.5 px-3",
        )}
        title={collapsed ? "New Chat" : undefined}
      >
        <span className="material-symbols-outlined text-[18px]">add</span>
        {!collapsed && "New Chat"}
      </button>

      {/* Nav */}
      <nav className="space-y-1">
        {NAV_ITEMS.map((item) => {
          const isActive = item.matchPrefix
            ? pathname.startsWith(item.href)
            : pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              title={collapsed ? item.label : undefined}
              className={cn(
                "flex items-center gap-2.5 rounded-md transition-all duration-200",
                collapsed ? "justify-center p-2" : "px-2.5 py-1.5",
                isActive
                  ? "text-[#ba9eff] bg-[#0a1020] font-bold"
                  : "text-[#dee5ff]/50 hover:text-[#dee5ff] hover:bg-[#1f2b49]/50",
              )}
            >
              <span
                className="material-symbols-outlined text-[18px]"
                style={isActive ? { fontVariationSettings: "'FILL' 1" } : undefined}
              >
                {item.icon}
              </span>
              {!collapsed && <span className="font-medium text-[13px]">{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      {/* Session history */}
      {!collapsed && (
        <div className="flex-1 overflow-y-auto min-h-0 mt-4 no-scrollbar">
          <Suspense
            fallback={
              <div className="px-4 py-2 text-xs text-[#a3aac4]">Loading...</div>
            }
          >
            <SidebarHistory />
          </Suspense>
        </div>
      )}

      {/* Spacer when collapsed */}
      {collapsed && <div className="flex-1" />}

      {/* Bottom */}
      <div className="pt-4 border-t border-[#40485d]/8 space-y-1 mt-auto">
        <SidebarLink href="/logs" icon="terminal" label="Logs" collapsed={collapsed} />
        <SidebarLink href="/help" icon="info" label="Info" collapsed={collapsed} />
      </div>
    </aside>
  );
}

function SidebarLink({
  href,
  icon,
  label,
  collapsed,
}: {
  href: string;
  icon: string;
  label: string;
  collapsed: boolean;
}) {
  const pathname = usePathname();
  const isActive = pathname === href;

  return (
    <Link
      href={href}
      title={collapsed ? label : undefined}
      className={cn(
        "flex items-center gap-2.5 rounded-md transition-all duration-200",
        collapsed ? "justify-center p-2" : "px-2.5 py-1.5",
        isActive
          ? "text-[#ba9eff] bg-[#0a1020] font-bold"
          : "text-[#dee5ff]/50 hover:text-[#dee5ff] hover:bg-[#1f2b49]/50",
      )}
    >
      <span
        className="material-symbols-outlined text-[18px]"
        style={isActive ? { fontVariationSettings: "'FILL' 1" } : undefined}
      >
        {icon}
      </span>
      {!collapsed && <span className="text-[13px]">{label}</span>}
    </Link>
  );
}
