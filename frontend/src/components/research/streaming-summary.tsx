"use client";

import { useEffect, useRef } from "react";

interface StreamingSummaryProps {
  content: string;
  isStreaming: boolean;
}

export function StreamingSummary({ content, isStreaming }: StreamingSummaryProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current && isStreaming) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [isStreaming]);

  if (!content && !isStreaming) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Summary will appear here as research progresses...
      </div>
    );
  }

  return (
    <div ref={containerRef} className="prose prose-invert prose-sm max-w-none">
      <div className="whitespace-pre-wrap">{content}</div>
      {isStreaming && (
        <span className="inline-block h-4 w-0.5 bg-indigo-400 animate-pulse ml-0.5" />
      )}
    </div>
  );
}
