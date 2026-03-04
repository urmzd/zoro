"use client";

import { KnowledgeGraph } from "@/components/graph/knowledge-graph";

export default function KnowledgePage() {
  return (
    <main className="h-screen w-screen overflow-hidden">
      <KnowledgeGraph />
    </main>
  );
}
