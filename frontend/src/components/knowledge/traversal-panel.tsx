"use client";

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
  }, []);

  const hasActivity = messages.length > 0 || isStreaming;

  return (
    <aside className="flex flex-col h-full bg-[#050a18]/50 rounded-2xl border border-[#40485d]/10 overflow-hidden glass-panel">
      <div className="p-5 border-b border-[#40485d]/10 flex justify-between items-center bg-[#050a18]/80">
        <h3 className="font-headline font-bold text-lg flex items-center gap-2">
          <span
            className="material-symbols-outlined text-[#ba9eff]"
            style={{ fontVariationSettings: "'FILL' 1" }}
          >
            auto_awesome
          </span>
          Traversal Agent
        </h3>
        {isStreaming && (
          <span className="text-[10px] uppercase tracking-widest text-[#ff716a] font-bold px-2 py-0.5 rounded-full bg-[#ff716a]/10">
            Streaming
          </span>
        )}
      </div>

      <div ref={scrollRef} className="flex-1 overflow-y-auto p-6 space-y-6">
        {hasActivity ? (
          <>
            {messages.map((msg, i) => (
              <ChatMessage key={`km-${i}`} message={msg} />
            ))}
            {isStreaming && (
              <StreamingMessage content={currentAssistantContent} toolCalls={currentToolCalls} />
            )}
            {error && (
              <div className="rounded-xl glass-card px-4 py-3 text-sm text-[#ff6e84] border border-[#ff6e84]/20">
                {error}
              </div>
            )}
          </>
        ) : (
          <div className="flex h-full flex-col items-center justify-center text-center text-[#a3aac4]">
            <span className="material-symbols-outlined text-4xl mb-3 opacity-40">hub</span>
            <p className="text-sm">Search to start exploring your knowledge graph</p>
            <p className="text-xs mt-1 opacity-60">
              You&apos;ll see the AI traversing nodes in real time
            </p>
          </div>
        )}
      </div>

      {hasActivity && (
        <div className="p-4 bg-[#050a18] border-t border-[#40485d]/10">
          <ChatInput
            onSend={onSend}
            onStop={onStop}
            isStreaming={isStreaming}
            placeholder="Refine your search..."
          />
        </div>
      )}
    </aside>
  );
}
