"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

export function SidebarKnowledge() {
  const pathname = usePathname();
  const isActive = pathname === "/knowledge";

  return (
    <div className="px-2 py-1">
      <Link
        href="/knowledge"
        className={cn(
          "flex items-center gap-2.5 rounded-md px-2.5 py-1.5 transition-all duration-200",
          isActive
            ? "text-[#ba9eff] bg-[#0f1930] font-bold"
            : "text-[#dee5ff]/60 hover:text-[#dee5ff] hover:bg-[#1f2b49]",
        )}
      >
        <span
          className="material-symbols-outlined text-[18px]"
          style={isActive ? { fontVariationSettings: "'FILL' 1" } : undefined}
        >
          hub
        </span>
        Knowledge Graph
      </Link>
    </div>
  );
}
