"use client";

import { cn } from "@/lib/utils";
import type { ChatMessage as ChatMessageType } from "@/app/lib/types";
import { ToolCallCard } from "./tool-call-card";

interface ChatMessageProps {
  message: ChatMessageType;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === "user";

  return (
    <div className={cn("flex gap-3", isUser && "justify-end")}>
      <div
        className={cn(
          "max-w-[80%] rounded-2xl px-4 py-3",
          isUser
            ? "bg-indigo-600 text-white"
            : "bg-muted/50 text-foreground",
        )}
      >
        {/* Tool calls rendered before text for assistant */}
        {!isUser && message.toolCalls?.map((tc) => (
          <ToolCallCard key={tc.id} toolCall={tc} />
        ))}

        {message.content && (
          <div className="prose prose-invert prose-sm max-w-none whitespace-pre-wrap">
            {message.content}
          </div>
        )}
      </div>
    </div>
  );
}
