"use client";

import { code } from "@streamdown/code";
import { useMemo } from "react";
import { Streamdown } from "streamdown";
import type { ToolCall } from "@/app/lib/types";
import { injectCitationLinks } from "./citation-text";
import { ToolCallCard } from "./tool-call-card";

interface StreamingMessageProps {
  content: string;
  toolCalls: ToolCall[];
}

// biome-ignore lint/suspicious/noExplicitAny: streamdown plugin types
const plugins = { code } as any;

export function StreamingMessage({ content, toolCalls }: StreamingMessageProps) {
  const hasContent = content || toolCalls.length > 0;

  const citedContent = useMemo(
    () => (content ? injectCitationLinks(content, toolCalls) : ""),
    [content, toolCalls],
  );

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

      {!hasContent ? (
        <div className="flex items-center gap-1.5 px-4 py-3 rounded-xl bg-[#050a18] border border-[#40485d]/5">
          <div
            className="w-2 h-2 rounded-full bg-[#ba9eff] typing-pulse"
            style={{ animationDelay: "0s" }}
          />
          <div
            className="w-2 h-2 rounded-full bg-[#ba9eff] typing-pulse"
            style={{ animationDelay: "0.2s" }}
          />
          <div
            className="w-2 h-2 rounded-full bg-[#ba9eff] typing-pulse"
            style={{ animationDelay: "0.4s" }}
          />
          <span className="ml-2 text-xs text-[#a3aac4] italic">Zoro is synthesizing data...</span>
        </div>
      ) : (
        <>
          {toolCalls.length > 0 && (
            <div className="w-full space-y-3">
              {toolCalls.map((tc) => (
                <ToolCallCard key={tc.id} toolCall={tc} />
              ))}
            </div>
          )}

          {citedContent && (
            <div className="max-w-[90%] text-[#dee5ff] leading-relaxed font-light">
              <Streamdown plugins={plugins} caret="block" isAnimating>
                {citedContent}
              </Streamdown>
            </div>
          )}
        </>
      )}
    </div>
  );
}
