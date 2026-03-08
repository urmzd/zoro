"use client";

import { code } from "@streamdown/code";
import { Streamdown } from "streamdown";
import type { ChatMessage as ChatMessageType } from "@/app/lib/types";
import { cn } from "@/lib/utils";
import { ToolCallCard } from "./tool-call-card";

interface ChatMessageProps {
  message: ChatMessageType;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const plugins = { code } as any;

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === "user";

  return (
    <div className={cn("flex gap-3", isUser && "justify-end")}>
      <div
        className={cn(
          "max-w-[80%] rounded-2xl px-4 py-3",
          isUser ? "bg-indigo-600 text-white" : "bg-muted/50 text-foreground",
        )}
      >
        {/* Tool calls rendered before text for assistant */}
        {!isUser && message.toolCalls?.map((tc) => <ToolCallCard key={tc.id} toolCall={tc} />)}

        {message.content &&
          (isUser ? (
            <div className="whitespace-pre-wrap">{message.content}</div>
          ) : (
            <Streamdown mode="static" plugins={plugins}>
              {message.content}
            </Streamdown>
          ))}
      </div>
    </div>
  );
}
