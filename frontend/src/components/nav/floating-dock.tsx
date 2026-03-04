"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { IconSearch, IconGraph, IconHome } from "@tabler/icons-react";

interface DockItem {
  title: string;
  icon: React.ReactNode;
  href: string;
}

const items: DockItem[] = [
  { title: "Home", icon: <IconHome className="h-5 w-5" />, href: "/" },
  { title: "Knowledge", icon: <IconGraph className="h-5 w-5" />, href: "/knowledge" },
];

export function FloatingDock({ className }: { className?: string }) {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null);

  return (
    <div
      className={cn(
        "fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-2 rounded-full border border-border/50 bg-background/80 backdrop-blur-md px-4 py-2 shadow-lg",
        className
      )}
    >
      {items.map((item, idx) => (
        <Link
          key={item.title}
          href={item.href}
          className="relative flex items-center justify-center rounded-full p-2 text-muted-foreground transition-colors hover:text-foreground"
          onMouseEnter={() => setHoveredIdx(idx)}
          onMouseLeave={() => setHoveredIdx(null)}
        >
          <AnimatePresence>
            {hoveredIdx === idx && (
              <motion.span
                className="absolute -top-8 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-md bg-foreground px-2 py-0.5 text-xs text-background"
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: 2 }}
              >
                {item.title}
              </motion.span>
            )}
          </AnimatePresence>
          <motion.div
            whileHover={{ scale: 1.2 }}
            transition={{ type: "spring", stiffness: 400, damping: 17 }}
          >
            {item.icon}
          </motion.div>
        </Link>
      ))}
    </div>
  );
}
