"use client";

import { useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { ResearchView } from "@/components/research/research-view";

function ResearchContent() {
  const searchParams = useSearchParams();
  const id = searchParams.get("id") || "";
  const query = searchParams.get("q") || "";

  if (!id) {
    return (
      <div className="flex h-screen items-center justify-center text-muted-foreground">
        Start a research query from the home page.
      </div>
    );
  }

  return <ResearchView sessionId={id} query={query} />;
}

export default function ResearchPage() {
  return (
    <Suspense>
      <ResearchContent />
    </Suspense>
  );
}
