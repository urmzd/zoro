"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { createChatSession } from "@/app/lib/api";

export function GlobalShortcuts() {
  const router = useRouter();

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const mod = e.metaKey || e.ctrlKey;
      if (!mod) return;

      if (e.key === "n") {
        e.preventDefault();
        createChatSession()
          .then(({ id }) => router.push(`/chat?id=${id}`))
          .catch(() => {});
      }

      if (e.key === "k") {
        e.preventDefault();
        // TODO: global search
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [router]);

  return null;
}
