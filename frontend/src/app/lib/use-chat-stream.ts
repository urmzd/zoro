"use client";

import { useCallback, useEffect, useRef } from "react";
import { useChatStore } from "@/lib/stores/chat-store";
import { getChatSession, sendChatMessage } from "./api";

export function useChatStream(sessionId: string | null) {
  const store = useChatStore();
  const abortRef = useRef<AbortController | null>(null);
  const loadedSessionRef = useRef<string | null>(null);
  // Capture stable action references via the Zustand store API
  const actionsRef = useRef(useChatStore.getState());
  actionsRef.current = useChatStore.getState();

  // Load existing messages when navigating to a previous session
  useEffect(() => {
    if (!sessionId || loadedSessionRef.current === sessionId) return;
    loadedSessionRef.current = sessionId;
    actionsRef.current.reset();

    getChatSession(sessionId)
      .then((session) => {
        if (session.messages?.length > 0) {
          actionsRef.current.loadMessages(session.messages);
        }
      })
      .catch(() => {
        // New session or fetch failed — start fresh
      });
  }, [sessionId]);

  const send = useCallback(
    async (content: string) => {
      if (!sessionId || !content.trim()) return;

      const actions = actionsRef.current;
      actions.addUserMessage(content);

      const controller = new AbortController();
      abortRef.current = controller;

      try {
        const resp = await sendChatMessage(sessionId, content);
        if (!resp.ok || !resp.body) {
          actions.setError("Failed to send message");
          return;
        }

        const reader = resp.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;
          if (controller.signal.aborted) {
            reader.cancel();
            break;
          }

          buffer += decoder.decode(value, { stream: true });

          // Process complete SSE messages (terminated by double newline)
          let idx: number;
          while ((idx = buffer.indexOf("\n\n")) !== -1) {
            const block = buffer.slice(0, idx);
            buffer = buffer.slice(idx + 2);

            let eventType = "";
            let dataStr = "";

            for (const line of block.split("\n")) {
              if (line.startsWith("event: ")) {
                eventType = line.slice(7).trim();
              } else if (line.startsWith("data: ")) {
                dataStr = line.slice(6);
              }
            }

            if (eventType && dataStr !== undefined) {
              try {
                processEvent(eventType, dataStr, actionsRef.current);
              } catch {
                // ignore parse errors
              }
            }
          }
        }

        // Process remaining buffer
        if (buffer.trim()) {
          let eventType = "";
          let dataStr = "";
          for (const line of buffer.split("\n")) {
            if (line.startsWith("event: ")) {
              eventType = line.slice(7).trim();
            } else if (line.startsWith("data: ")) {
              dataStr = line.slice(6);
            }
          }
          if (eventType) {
            try {
              processEvent(eventType, dataStr, actionsRef.current);
            } catch {
              // ignore
            }
          }
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          actionsRef.current.setError(
            err instanceof Error ? err.message : "Connection error",
          );
        }
      } finally {
        abortRef.current = null;
      }
    },
    [sessionId],
  );

  const stop = useCallback(() => {
    abortRef.current?.abort();
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

export function processEvent(type: string, dataStr: string, actions: StoreActions) {
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
