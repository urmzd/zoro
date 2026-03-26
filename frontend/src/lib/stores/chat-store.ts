"use client";

import { create } from "zustand";
import type { ChatMessage, ChatState, ToolCall } from "@/app/lib/types";

function msgId(): string {
  return crypto.randomUUID();
}

interface ChatStore extends ChatState {
  pendingQuery: string | null;
  setPendingQuery: (query: string | null) => void;
  addUserMessage: (content: string) => void;
  appendAssistantContent: (content: string) => void;
  addToolCallStart: (toolCall: ToolCall) => void;
  setToolCallResult: (id: string, result: string) => void;
  finalizeTurn: () => void;
  setStatus: (status: ChatState["status"]) => void;
  setError: (error: string) => void;
  loadMessages: (messages: ChatMessage[]) => void;
  reset: () => void;
}

const initialState: ChatState = {
  messages: [],
  currentAssistantContent: "",
  currentToolCalls: [],
  status: "idle",
  error: null,
};

export const useChatStore = create<ChatStore>((set) => ({
  ...initialState,
  pendingQuery: null,

  setPendingQuery: (query) => set({ pendingQuery: query }),

  addUserMessage: (content) =>
    set((s) => ({
      messages: [...s.messages, { id: msgId(), role: "user", content }],
      currentAssistantContent: "",
      currentToolCalls: [],
      status: "streaming",
      error: null,
    })),

  appendAssistantContent: (content) =>
    set((s) => ({
      currentAssistantContent: s.currentAssistantContent + content,
    })),

  addToolCallStart: (toolCall) =>
    set((s) => ({
      currentToolCalls: [...s.currentToolCalls, toolCall],
    })),

  setToolCallResult: (id, result) =>
    set((s) => ({
      currentToolCalls: s.currentToolCalls.map((tc) => (tc.id === id ? { ...tc, result } : tc)),
    })),

  finalizeTurn: () =>
    set((s) => {
      const assistantMsg: ChatMessage = {
        id: msgId(),
        role: "assistant",
        content: s.currentAssistantContent,
        toolCalls: s.currentToolCalls.length > 0 ? s.currentToolCalls : undefined,
      };
      return {
        messages: [...s.messages, assistantMsg],
        currentAssistantContent: "",
        currentToolCalls: [],
        status: "complete",
      };
    }),

  setStatus: (status) => set({ status }),

  setError: (error) => set({ status: "error", error }),

  loadMessages: (messages) =>
    set({
      messages: messages.map((m) => ({ ...m, id: m.id || msgId() })),
      currentAssistantContent: "",
      currentToolCalls: [],
      status: "idle",
      error: null,
    }),

  reset: () => set(initialState),
}));
