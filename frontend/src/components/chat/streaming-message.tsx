"use client";

import type { ToolCall } from "@/app/lib/types";
import { Streamdown } from "streamdown";
import { code } from "@streamdown/code";
import { ToolCallCard } from "./tool-call-card";

interface StreamingMessageProps {
  content: string;
  toolCalls: ToolCall[];
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const plugins = { code } as any;

export function StreamingMessage({ content, toolCalls }: StreamingMessageProps) {
  const hasContent = content || toolCalls.length > 0;

  if (!hasContent) {
    return (
      <div className="flex gap-3">
        <div className="bg-muted/50 rounded-2xl px-4 py-3">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span className="h-2 w-2 rounded-full bg-indigo-400 animate-pulse" />
            Thinking...
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex gap-3">
      <div className="max-w-[80%] bg-muted/50 rounded-2xl px-4 py-3 text-foreground">
        {toolCalls.map((tc) => (
          <ToolCallCard key={tc.id} toolCall={tc} />
        ))}

        {content && (
          <Streamdown plugins={plugins} caret="block" isAnimating>
            {content}
          </Streamdown>
        )}
      </div>
    </div>
  );
}
