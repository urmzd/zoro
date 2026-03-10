"use client";

import { useParams, useSearchParams } from "next/navigation";
import { ResearchView } from "@/components/research/research-view";

export default function ResearchPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const id = params.id as string;
  const query = searchParams.get("q") || "";

  return <ResearchView sessionId={id} query={query} />;
}
