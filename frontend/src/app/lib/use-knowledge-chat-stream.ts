"use client";

import { useCallback, useEffect, useRef } from "react";
import { useKnowledgeChatStore } from "@/lib/stores/knowledge-chat-store";
import { createChatSession, type SSEEvent, sendMessageSSE } from "./api";
import { processEvent, type StoreActions } from "./use-chat-stream";

interface UseKnowledgeChatStreamOptions {
  onToolCallResult?: (name: string, result: string) => void;
}

export function useKnowledgeChatStream(options: UseKnowledgeChatStreamOptions = {}) {
  const store = useKnowledgeChatStore();
  const cancelRef = useRef<(() => void) | null>(null);
  const actionsRef = useRef(useKnowledgeChatStore.getState());
  actionsRef.current = useKnowledgeChatStore.getState();
  const optionsRef = useRef(options);
  optionsRef.current = options;

  // Cleanup listener on unmount
  useEffect(() => {
    return () => {
      if (cancelRef.current) {
        cancelRef.current();
        cancelRef.current = null;
      }
    };
  }, []);

  const send = useCallback(async (content: string) => {
    if (!content.trim()) return;

    const actions = actionsRef.current;

    // Auto-create session if needed
    let sid = actions.sessionId;
    if (!sid) {
      try {
        const session = await createChatSession();
        sid = session.id;
        actions.setSessionId(session.id);
      } catch {
        actions.setError("Failed to create session");
        return;
      }
    }

    actions.addUserMessage(content);

    // Cleanup previous connection
    if (cancelRef.current) {
      cancelRef.current();
      cancelRef.current = null;
    }

    // Wrap store actions to intercept tool_call_result for the callback
    const wrappedActions: StoreActions = {
      appendAssistantContent: actions.appendAssistantContent,
      addToolCallStart: actions.addToolCallStart,
      setToolCallResult: (id: string, result: string) => {
        actions.setToolCallResult(id, result);
        const current = useKnowledgeChatStore.getState().currentToolCalls;
        const tc = current.find((t) => t.id === id);
        if (tc && optionsRef.current.onToolCallResult) {
          optionsRef.current.onToolCallResult(tc.name, result);
        }
      },
      finalizeTurn: actions.finalizeTurn,
      setError: actions.setError,
    };

    // Stream SSE events from the HTTP server
    cancelRef.current = sendMessageSSE(
      sid,
      content,
      (event: SSEEvent) => {
        processEvent(event.type, JSON.stringify(event.data), wrappedActions);
      },
      (err) => {
        actionsRef.current.setError(err.message);
      },
    );
  }, []);

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
