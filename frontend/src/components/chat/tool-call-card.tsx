"use client";

import { IconBrain, IconChevronDown, IconDatabase, IconSearch } from "@tabler/icons-react";
import { useState } from "react";
import type { ToolCall } from "@/app/lib/types";
import { cn } from "@/lib/utils";

const TOOL_CONFIG: Record<string, { icon: typeof IconSearch; color: string; label: string }> = {
  web_search: {
    icon: IconSearch,
    color: "text-blue-400 bg-blue-500/10 border-blue-500/20",
    label: "Web Search",
  },
  search_knowledge: {
    icon: IconBrain,
    color: "text-purple-400 bg-purple-500/10 border-purple-500/20",
    label: "Knowledge Search",
  },
  store_knowledge: {
    icon: IconDatabase,
    color: "text-green-400 bg-green-500/10 border-green-500/20",
    label: "Store Knowledge",
  },
};

interface ToolCallCardProps {
  toolCall: ToolCall;
}

export function ToolCallCard({ toolCall }: ToolCallCardProps) {
  const [expanded, setExpanded] = useState(false);
  const config = TOOL_CONFIG[toolCall.name] || TOOL_CONFIG.web_search;
  const Icon = config.icon;

  // Extract a preview from arguments
  let preview = "";
  try {
    const args = JSON.parse(toolCall.arguments);
    preview = args.query || args.text?.slice(0, 60) || toolCall.arguments;
  } catch {
    preview = toolCall.arguments;
  }

  const isLoading = !toolCall.result;

  return (
    <div className={cn("rounded-lg border text-sm my-2", config.color)}>
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-2 px-3 py-2"
      >
        <Icon className="h-4 w-4 shrink-0" />
        <span className="font-medium">{config.label}</span>
        <span className="truncate text-xs opacity-70 flex-1 text-left">{preview}</span>
        {isLoading && (
          <span className="h-3 w-3 rounded-full border-2 border-current border-t-transparent animate-spin shrink-0" />
        )}
        <IconChevronDown
          className={cn("h-3.5 w-3.5 shrink-0 transition-transform", expanded && "rotate-180")}
        />
      </button>
      {expanded && (
        <div className="border-t border-inherit px-3 py-2 space-y-2">
          <div>
            <div className="text-xs font-medium opacity-60 mb-0.5">Arguments</div>
            <pre className="text-xs whitespace-pre-wrap break-all opacity-80">
              {toolCall.arguments}
            </pre>
          </div>
          {toolCall.result && (
            <div>
              <div className="text-xs font-medium opacity-60 mb-0.5">Result</div>
              <pre className="text-xs whitespace-pre-wrap break-all opacity-80 max-h-40 overflow-y-auto">
                {toolCall.result}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
