"use client";

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

  useEffect(() => {
    if (sentInitial.current || !sessionId) return;
    const pending = useChatStore.getState().pendingQuery;
    if (!pending) return;
    sentInitial.current = true;
    useChatStore.getState().setPendingQuery(null);
    send(pending);
  }, [sessionId, send]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: triggers scroll when messages/streaming content changes
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, currentAssistantContent, currentToolCalls]);

  const isStreaming = status === "streaming";

  return (
    <div className="flex flex-col h-full">
      {/* Tab toggle */}
      <div className="flex items-center px-8 h-12 shrink-0">
        <div className="flex items-center p-1 bg-[#050a18] rounded-xl">
          <button
            type="button"
            onClick={() => setActiveTab("chat")}
            className={cn(
              "px-4 py-1.5 text-xs font-bold rounded-lg transition-colors",
              activeTab === "chat"
                ? "bg-[#0f1930] text-[#ba9eff]"
                : "text-[#a3aac4] hover:text-[#dee5ff]",
            )}
          >
            Chat
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("graph")}
            className={cn(
              "px-4 py-1.5 text-xs font-medium rounded-lg transition-colors",
              activeTab === "graph"
                ? "bg-[#0f1930] text-[#ba9eff]"
                : "text-[#a3aac4] hover:text-[#dee5ff]",
            )}
          >
            Knowledge Graph
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 min-h-0">
        {activeTab === "chat" ? (
          <div className="flex flex-col h-full">
            {/* Messages */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto px-8 py-12">
              <div className="max-w-5xl mx-auto w-full space-y-12">
                {messages.map((msg) => (
                  <ChatMessage key={msg.id} message={msg} />
                ))}

                {isStreaming && (
                  <StreamingMessage
                    content={currentAssistantContent}
                    toolCalls={currentToolCalls}
                  />
                )}

                {error && (
                  <div className="rounded-xl glass-card px-4 py-3 text-sm text-[#ff6e84] border border-[#ff6e84]/20">
                    {error}
                  </div>
                )}
              </div>
            </div>

            {/* Input */}
            <div className="p-8 pt-0">
              <div className="max-w-4xl mx-auto">
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
