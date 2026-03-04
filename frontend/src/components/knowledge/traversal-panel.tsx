"use client";

import { IconBrain } from "@tabler/icons-react";
import { useEffect, useRef } from "react";
import type { ChatMessage as ChatMessageType, ToolCall } from "@/app/lib/types";
import { ChatInput } from "@/components/chat/chat-input";
import { ChatMessage } from "@/components/chat/chat-message";
import { StreamingMessage } from "@/components/chat/streaming-message";

interface TraversalPanelProps {
  messages: ChatMessageType[];
  currentAssistantContent: string;
  currentToolCalls: ToolCall[];
  isStreaming: boolean;
  error: string | null;
  onSend: (content: string) => void;
  onStop: () => void;
}

export function TraversalPanel({
  messages,
  currentAssistantContent,
  currentToolCalls,
  isStreaming,
  error,
  onSend,
  onStop,
}: TraversalPanelProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, currentAssistantContent, currentToolCalls]);

  const hasActivity = messages.length > 0 || isStreaming;

  return (
    <div className="flex h-full flex-col">
      {/* Scrollable messages area */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4">
        {hasActivity ? (
          <div className="space-y-4">
            {messages.map((msg, i) => (
              <ChatMessage key={`km-${i}`} message={msg} />
            ))}

            {isStreaming && (
              <StreamingMessage
                content={currentAssistantContent}
                toolCalls={currentToolCalls}
              />
            )}

            {error && (
              <div className="rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-2 text-sm text-destructive">
                {error}
              </div>
            )}
          </div>
        ) : (
          <div className="flex h-full flex-col items-center justify-center text-center text-muted-foreground">
            <IconBrain className="h-10 w-10 mb-3 opacity-40" />
            <p className="text-sm">Search to start exploring your knowledge graph</p>
            <p className="text-xs mt-1 opacity-60">
              You'll see the AI traversing nodes in real time
            </p>
          </div>
        )}
      </div>

      {/* Follow-up input */}
      {hasActivity && (
        <div className="border-t border-border px-4 py-3">
          <ChatInput
            onSend={onSend}
            onStop={onStop}
            isStreaming={isStreaming}
            placeholder="Refine your search..."
          />
        </div>
      )}
    </div>
  );
}
