"use client";

import { useCallback, useRef } from "react";
import { useKnowledgeChatStore } from "@/lib/stores/knowledge-chat-store";
import { createChatSession, sendChatMessage } from "./api";
import { processEvent, type StoreActions } from "./use-chat-stream";

interface UseKnowledgeChatStreamOptions {
  onToolCallResult?: (name: string, result: string) => void;
}

export function useKnowledgeChatStream(
  options: UseKnowledgeChatStreamOptions = {},
) {
  const store = useKnowledgeChatStore();
  const abortRef = useRef<AbortController | null>(null);
  const actionsRef = useRef(useKnowledgeChatStore.getState());
  actionsRef.current = useKnowledgeChatStore.getState();
  const optionsRef = useRef(options);
  optionsRef.current = options;

  const send = useCallback(async (content: string) => {
    if (!content.trim()) return;

    const actions = actionsRef.current;

    // Auto-create session if needed
    let sid = actions.sessionId;
    if (!sid) {
      try {
        const { id } = await createChatSession();
        sid = id;
        actions.setSessionId(id);
      } catch {
        actions.setError("Failed to create session");
        return;
      }
    }

    actions.addUserMessage(content);

    const controller = new AbortController();
    abortRef.current = controller;

    // Wrap store actions to intercept tool_call_result for the callback
    const wrappedActions: StoreActions = {
      appendAssistantContent: actions.appendAssistantContent,
      addToolCallStart: actions.addToolCallStart,
      setToolCallResult: (id: string, result: string) => {
        actions.setToolCallResult(id, result);
        // Find the tool call to get the name
        const current = useKnowledgeChatStore.getState().currentToolCalls;
        const tc = current.find((t) => t.id === id);
        if (tc && optionsRef.current.onToolCallResult) {
          optionsRef.current.onToolCallResult(tc.name, result);
        }
      },
      finalizeTurn: actions.finalizeTurn,
      setError: actions.setError,
    };

    try {
      const resp = await sendChatMessage(sid, content);
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
              processEvent(eventType, dataStr, wrappedActions);
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
            processEvent(eventType, dataStr, wrappedActions);
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
  }, []);

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
