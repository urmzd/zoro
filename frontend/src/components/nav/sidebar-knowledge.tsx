"use client";

import { IconGraph } from "@tabler/icons-react";
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
          "flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
          isActive
            ? "bg-muted text-foreground"
            : "text-muted-foreground hover:text-foreground hover:bg-muted/50",
        )}
      >
        <IconGraph className="h-4 w-4" />
        Knowledge Graph
      </Link>
    </div>
  );
}
