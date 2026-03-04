"use client";

import { create } from "zustand";
import type { ChatMessage, ChatState, ToolCall } from "@/app/lib/types";

interface KnowledgeChatStore extends ChatState {
  sessionId: string | null;
  setSessionId: (id: string) => void;
  addUserMessage: (content: string) => void;
  appendAssistantContent: (content: string) => void;
  addToolCallStart: (toolCall: ToolCall) => void;
  setToolCallResult: (id: string, result: string) => void;
  finalizeTurn: () => void;
  setStatus: (status: ChatState["status"]) => void;
  setError: (error: string) => void;
  reset: () => void;
}

const initialState: ChatState & { sessionId: string | null } = {
  sessionId: null,
  messages: [],
  currentAssistantContent: "",
  currentToolCalls: [],
  status: "idle",
  error: null,
};

export const useKnowledgeChatStore = create<KnowledgeChatStore>((set) => ({
  ...initialState,

  setSessionId: (id) => set({ sessionId: id }),

  addUserMessage: (content) =>
    set((s) => ({
      messages: [...s.messages, { role: "user", content }],
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
      currentToolCalls: s.currentToolCalls.map((tc) =>
        tc.id === id ? { ...tc, result } : tc,
      ),
    })),

  finalizeTurn: () =>
    set((s) => {
      const assistantMsg: ChatMessage = {
        role: "assistant",
        content: s.currentAssistantContent,
        toolCalls:
          s.currentToolCalls.length > 0 ? s.currentToolCalls : undefined,
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

  reset: () => set(initialState),
}));
