"use client";

import { IconGraph, IconMessage } from "@tabler/icons-react";
import { useEffect, useRef, useState } from "react";
import { useChatStream } from "@/app/lib/use-chat-stream";
import { GraphErrorBoundary } from "@/components/graph/graph-error-boundary";
import { KnowledgeGraph } from "@/components/graph/knowledge-graph";
import { useChatStore } from "@/lib/stores/chat-store";
import { cn } from "@/lib/utils";
import { ChatInput } from "./chat-input";
import { ChatMessage } from "./chat-message";
import { StreamingMessage } from "./streaming-message";

interface ChatViewProps {
  sessionId: string;
}

type Tab = "chat" | "graph";

export function ChatView({ sessionId }: ChatViewProps) {
  const { messages, currentAssistantContent, currentToolCalls, status, error, send, stop } =
    useChatStream(sessionId);

  const scrollRef = useRef<HTMLDivElement>(null);
  const sentInitial = useRef(false);
  const [activeTab, setActiveTab] = useState<Tab>("chat");

  // Consume pending query from store (set by landing page before navigation)
  useEffect(() => {
    if (sentInitial.current || !sessionId) return;
    const pending = useChatStore.getState().pendingQuery;
    if (!pending) return;
    sentInitial.current = true;
    useChatStore.getState().setPendingQuery(null);
    send(pending);
  }, [sessionId, send]);

  // Auto-scroll on new content
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, currentAssistantContent, currentToolCalls]);

  const isStreaming = status === "streaming";

  return (
    <div className="flex flex-col h-full">
      {/* Tab bar */}
      <div className="flex items-center gap-1 px-4 py-2 border-b border-border/50">
        <button
          type="button"
          onClick={() => setActiveTab("chat")}
          className={cn(
            "flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors",
            activeTab === "chat"
              ? "bg-muted text-foreground"
              : "text-muted-foreground hover:text-foreground hover:bg-muted/50",
          )}
        >
          <IconMessage className="h-4 w-4" />
          Chat
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("graph")}
          className={cn(
            "flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors",
            activeTab === "graph"
              ? "bg-muted text-foreground"
              : "text-muted-foreground hover:text-foreground hover:bg-muted/50",
          )}
        >
          <IconGraph className="h-4 w-4" />
          Knowledge Graph
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 min-h-0">
        {activeTab === "chat" ? (
          <div className="flex flex-col h-full">
            {/* Message thread */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-6">
              <div className="max-w-3xl mx-auto space-y-4">
                {messages.map((msg, i) => (
                  <ChatMessage key={`msg-${i}`} message={msg} />
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
            </div>

            {/* Input */}
            <div className="border-t border-border px-4 py-3">
              <div className="max-w-3xl mx-auto">
                <ChatInput onSend={send} onStop={stop} isStreaming={isStreaming} />
              </div>
            </div>
          </div>
        ) : (
          <GraphErrorBoundary>
            <KnowledgeGraph />
          </GraphErrorBoundary>
        )}
      </div>
    </div>
  );
}
