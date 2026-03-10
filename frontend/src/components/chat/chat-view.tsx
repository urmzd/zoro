"use client";

import { useEffect, useRef } from "react";
import { useChatStream } from "@/app/lib/use-chat-stream";
import { ChatInput } from "./chat-input";
import { ChatMessage } from "./chat-message";
import { GraphSidePanel } from "./graph-side-panel";
import { StreamingMessage } from "./streaming-message";

interface ChatViewProps {
  sessionId: string;
  initialQuery?: string;
}

export function ChatView({ sessionId, initialQuery }: ChatViewProps) {
  const { messages, currentAssistantContent, currentToolCalls, status, error, send, stop } =
    useChatStream(sessionId);

  const scrollRef = useRef<HTMLDivElement>(null);
  const sentInitial = useRef(false);

  // Auto-send initial query
  useEffect(() => {
    if (initialQuery && !sentInitial.current && sessionId) {
      sentInitial.current = true;
      send(initialQuery);
    }
  }, [initialQuery, sessionId, send]);

  // Auto-scroll on new content
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, currentAssistantContent, currentToolCalls]);

  const isStreaming = status === "streaming";

  return (
    <div className="flex h-screen">
      <div className="flex-1 flex flex-col min-w-0">
        {/* Message thread */}
        <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-6">
          <div className="max-w-3xl mx-auto space-y-4">
            {messages.map((msg, i) => (
              <ChatMessage key={`msg-${i}`} message={msg} />
            ))}

            {isStreaming && (
              <StreamingMessage content={currentAssistantContent} toolCalls={currentToolCalls} />
            )}

            {error && (
              <div className="rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-2 text-sm text-destructive">
                {error}
              </div>
            )}
          </div>
        </div>

        {/* Input */}
        <div className="border-t border-border px-4 py-3">
          <div className="max-w-3xl mx-auto">
            <ChatInput onSend={send} onStop={stop} isStreaming={isStreaming} />
          </div>
        </div>
      </div>

      <GraphSidePanel sessionId={sessionId} />
    </div>
  );
}
