"use client";

import { useCallback, useEffect, useRef } from "react";
import { useChatStore } from "@/lib/stores/chat-store";
import { getChatSession, sendMessageSSE, type SSEEvent } from "./api";

export function useChatStream(sessionId: string | null) {
  const store = useChatStore();
  const cancelRef = useRef<(() => void) | null>(null);
  const loadedSessionRef = useRef<string | null>(null);
  const actionsRef = useRef(useChatStore.getState());
  actionsRef.current = useChatStore.getState();

  // Load existing messages when navigating to a previous session
  useEffect(() => {
    if (!sessionId || loadedSessionRef.current === sessionId) return;
    loadedSessionRef.current = sessionId;
    actionsRef.current.reset();

    getChatSession(sessionId)
      .then((session) => {
        if (session.messages && session.messages.length > 0) {
          const mapped = session.messages.map((m) => ({
            role: m.role as "user" | "assistant" | "tool",
            content: m.content ?? "",
            toolCalls: m.toolCalls?.map((tc) => ({
              id: tc.id ?? "",
              name: tc.name ?? "",
              arguments: tc.arguments ?? "",
              result: tc.result,
            })),
          }));
          actionsRef.current.loadMessages(mapped);
        }
      })
      .catch(() => {
        // New session or fetch failed — start fresh
      });
  }, [sessionId]);

  // Cleanup listener on unmount
  useEffect(() => {
    return () => {
      if (cancelRef.current) {
        cancelRef.current();
        cancelRef.current = null;
      }
    };
  }, []);

  const send = useCallback(
    async (content: string) => {
      if (!sessionId || !content.trim()) return;

      const actions = actionsRef.current;
      actions.addUserMessage(content);

      // Cleanup previous connection
      if (cancelRef.current) {
        cancelRef.current();
        cancelRef.current = null;
      }

      // Stream SSE events from the HTTP server
      cancelRef.current = sendMessageSSE(
        sessionId,
        content,
        (event: SSEEvent) => {
          processEvent(
            event.type,
            JSON.stringify(event.data),
            actionsRef.current,
          );
        },
        (err) => {
          actionsRef.current.setError(err.message);
        },
      );
    },
    [sessionId],
  );

  const stop = useCallback(() => {
    if (cancelRef.current) {
      cancelRef.current();
      cancelRef.current = null;
    }
    actionsRef.current.finalizeTurn();
  }, []);

  return {
    messages: store.messages,
    currentAssistantContent: store.currentAssistantContent,
    currentToolCalls: store.currentToolCalls,
    status: store.status,
    error: store.error,
    send,
    stop,
  };
}

export interface StoreActions {
  appendAssistantContent: (content: string) => void;
  addToolCallStart: (toolCall: {
    id: string;
    name: string;
    arguments: string;
  }) => void;
  setToolCallResult: (id: string, result: string) => void;
  finalizeTurn: () => void;
  setError: (error: string) => void;
}

export function processEvent(
  type: string,
  dataStr: string,
  actions: StoreActions,
) {
  const data = dataStr === "null" ? null : JSON.parse(dataStr);

  switch (type) {
    case "text_delta":
      if (data?.content) {
        actions.appendAssistantContent(data.content);
      }
      break;
    case "tool_call_start":
      actions.addToolCallStart({
        id: data.id,
        name: data.name,
        arguments: data.arguments || "",
      });
      break;
    case "tool_call_result":
      actions.setToolCallResult(data.id, data.result || "");
      break;
    case "done":
      actions.finalizeTurn();
      break;
    case "error":
      actions.setError(data?.message || "Unknown error");
      break;
  }
}
