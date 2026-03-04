"use client";

import { IconSend, IconPlayerStop } from "@tabler/icons-react";
import { useRef, useState } from "react";

interface ChatInputProps {
  onSend: (content: string) => void;
  onStop: () => void;
  isStreaming: boolean;
  disabled?: boolean;
  placeholder?: string;
}

export function ChatInput({ onSend, onStop, isStreaming, disabled, placeholder = "Send a message..." }: ChatInputProps) {
  const [value, setValue] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  function handleSubmit() {
    if (!value.trim() || disabled) return;
    onSend(value.trim());
    setValue("");
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (isStreaming) return;
      handleSubmit();
    }
  }

  function handleInput() {
    const el = textareaRef.current;
    if (el) {
      el.style.height = "auto";
      el.style.height = `${Math.min(el.scrollHeight, 200)}px`;
    }
  }

  return (
    <div className="flex items-end gap-2 border border-border rounded-2xl bg-background/80 backdrop-blur-sm px-4 py-3">
      <textarea
        ref={textareaRef}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={handleKeyDown}
        onInput={handleInput}
        placeholder={placeholder}
        disabled={disabled}
        rows={1}
        className="flex-1 resize-none bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground text-sm max-h-[200px]"
      />
      {isStreaming ? (
        <button
          type="button"
          onClick={onStop}
          className="shrink-0 rounded-full bg-destructive p-2 text-destructive-foreground transition-opacity hover:opacity-80"
        >
          <IconPlayerStop className="h-4 w-4" />
        </button>
      ) : (
        <button
          type="button"
          onClick={handleSubmit}
          disabled={!value.trim() || disabled}
          className="shrink-0 rounded-full bg-indigo-600 p-2 text-white transition-opacity disabled:opacity-30 hover:opacity-80"
        >
          <IconSend className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}
