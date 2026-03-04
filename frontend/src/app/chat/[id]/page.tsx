"use client";

import { useParams, useSearchParams } from "next/navigation";
import { ChatView } from "@/components/chat/chat-view";

export default function ChatPage() {
  const params = useParams();
  const searchParams = useSearchParams();

  const sessionId = params.id as string;
  const initialQuery = searchParams.get("q") || undefined;

  return <ChatView sessionId={sessionId} initialQuery={initialQuery} />;
}
