"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { ChatView } from "@/components/chat/chat-view";

function ChatContent() {
  const searchParams = useSearchParams();

  const sessionId = searchParams.get("id") || "";
  const initialQuery = searchParams.get("q") || undefined;

  if (!sessionId) {
    return (
      <div className="flex h-screen items-center justify-center text-muted-foreground">
        Select a chat session or start a new one.
      </div>
    );
  }

  return <ChatView sessionId={sessionId} initialQuery={initialQuery} />;
}

export default function ChatPage() {
  return (
    <Suspense>
      <ChatContent />
    </Suspense>
  );
}
