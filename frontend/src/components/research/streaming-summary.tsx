"use client";

import { code } from "@streamdown/code";
import { useEffect, useRef } from "react";
import { Streamdown } from "streamdown";

interface StreamingSummaryProps {
  content: string;
  isStreaming: boolean;
}

// biome-ignore lint/suspicious/noExplicitAny: streamdown plugin types
const plugins = { code } as any;

export function StreamingSummary({ content, isStreaming }: StreamingSummaryProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: content triggers scroll on new streaming data
  useEffect(() => {
    if (containerRef.current && isStreaming) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [isStreaming, content]);

  if (!content && !isStreaming) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Summary will appear here as research progresses...
      </div>
    );
  }

  return (
    <div ref={containerRef} className="prose prose-invert prose-sm max-w-none overflow-y-auto">
      {isStreaming ? (
        <Streamdown plugins={plugins} caret="block" isAnimating>
          {content}
        </Streamdown>
      ) : (
        <Streamdown mode="static" plugins={plugins}>
          {content}
        </Streamdown>
      )}
    </div>
  );
}
