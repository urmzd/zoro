"use client";

import { useMemo } from "react";
import { code } from "@streamdown/code";
import { Streamdown } from "streamdown";
import type { ChatMessage as ChatMessageType } from "@/app/lib/types";
import { injectCitationLinks } from "./citation-text";
import { ToolCallCard } from "./tool-call-card";

interface ChatMessageProps {
  message: ChatMessageType;
}

// biome-ignore lint/suspicious/noExplicitAny: streamdown plugin types
const plugins = { code } as any;

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === "user";

  const citedContent = useMemo(
    () =>
      !isUser && message.content && message.toolCalls
        ? injectCitationLinks(message.content, message.toolCalls)
        : message.content,
    [isUser, message.content, message.toolCalls],
  );

  if (isUser) {
    return (
      <div className="flex flex-col items-end gap-3 group">
        <div className="max-w-[80%] bg-[#141f38] px-6 py-4 rounded-2xl rounded-tr-none shadow-sm text-[#dee5ff]/90 leading-relaxed border border-[#40485d]/10">
          <p className="whitespace-pre-wrap">{message.content}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-start gap-4">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg zoro-gradient-bg flex items-center justify-center">
          <span
            className="material-symbols-outlined text-xs text-black"
            style={{ fontVariationSettings: "'FILL' 1" }}
          >
            auto_awesome
          </span>
        </div>
        <span className="font-headline font-bold text-sm tracking-tight">Zoro</span>
      </div>

      {/* Tool calls */}
      {message.toolCalls && message.toolCalls.length > 0 && (
        <div className="w-full space-y-3">
          {message.toolCalls.map((tc) => (
            <ToolCallCard key={tc.id} toolCall={tc} />
          ))}
        </div>
      )}

      {/* Content */}
      {citedContent && (
        <div className="max-w-[90%] space-y-4 text-[#dee5ff] leading-relaxed font-light">
          <Streamdown mode="static" plugins={plugins}>
            {citedContent}
          </Streamdown>
        </div>
      )}
    </div>
  );
}
