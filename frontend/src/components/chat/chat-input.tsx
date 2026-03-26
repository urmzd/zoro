"use client";

import { useRef, useState } from "react";

interface ChatInputProps {
  onSend: (content: string) => void;
  onStop: () => void;
  isStreaming: boolean;
  disabled?: boolean;
  placeholder?: string;
}

export function ChatInput({
  onSend,
  onStop,
  isStreaming,
  disabled,
  placeholder = "Message Zoro...",
}: ChatInputProps) {
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
    <div className="relative group">
      <div className="absolute -inset-1 bg-gradient-to-r from-[#ba9eff]/8 via-[#699cff]/8 to-[#ba9eff]/8 rounded-[2rem] blur-xl opacity-0 group-focus-within:opacity-100 transition-opacity duration-500" />
      <div className="relative bg-black border border-[#40485d]/12 rounded-[2rem] p-4 flex flex-col gap-3 shadow-2xl">
        <div className="flex items-end gap-3 px-2">
          <textarea
            ref={textareaRef}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={handleKeyDown}
            onInput={handleInput}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            className="flex-1 bg-transparent border-none focus:ring-0 text-[#dee5ff] placeholder-[#a3aac4]/40 resize-none max-h-[200px] min-h-[48px] py-2 no-scrollbar"
          />
          <div className="flex items-center gap-2 mb-1">
            {isStreaming && (
              <button
                type="button"
                onClick={onStop}
                aria-label="Stop generating"
                className="w-10 h-10 rounded-full flex items-center justify-center text-[#ff6e84] border border-[#ff6e84]/20 hover:bg-[#ff6e84]/10 transition-colors"
              >
                <span className="material-symbols-outlined">stop_circle</span>
              </button>
            )}
            <button
              type="button"
              onClick={handleSubmit}
              disabled={!value.trim() || disabled || isStreaming}
              aria-label="Send message"
              className="w-10 h-10 rounded-full zoro-gradient-bg flex items-center justify-center text-black shadow-lg shadow-[#ba9eff]/20 hover:scale-105 active:scale-95 transition-all disabled:opacity-30"
            >
              <span className="material-symbols-outlined">arrow_upward</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
