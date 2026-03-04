"use client";

import { useSearchParams } from "next/navigation";
import { use } from "react";
import { ResearchView } from "@/components/research/research-view";

export default function ResearchPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const searchParams = useSearchParams();
  const query = searchParams.get("q") || "";

  return <ResearchView sessionId={id} query={query} />;
}
