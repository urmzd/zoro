"use client";

import { useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { ChatView } from "@/components/chat/chat-view";

function ChatContent() {
  const searchParams = useSearchParams();
  const sessionId = searchParams.get("id") || "";

  if (!sessionId) {
    return (
      <div className="flex h-full items-center justify-center text-muted-foreground">
        Select a chat session or start a new one.
      </div>
    );
  }

  return <ChatView sessionId={sessionId} />;
}

export default function ChatPage() {
  return (
    <Suspense>
      <ChatContent />
    </Suspense>
  );
}
