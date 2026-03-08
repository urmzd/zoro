"use client";

import { useCallback, useEffect, useRef } from "react";
import { listen, type UnlistenFn } from "@tauri-apps/api/event";
import { useKnowledgeChatStore } from "@/lib/stores/knowledge-chat-store";
import { createChatSession, sendChatMessage } from "./api";
import { processEvent, type StoreActions } from "./use-chat-stream";

interface SSEEvent {
  type: string;
  data: unknown;
}

interface UseKnowledgeChatStreamOptions {
  onToolCallResult?: (name: string, result: string) => void;
}

export function useKnowledgeChatStream(
  options: UseKnowledgeChatStreamOptions = {},
) {
  const store = useKnowledgeChatStore();
  const unlistenRef = useRef<UnlistenFn | null>(null);
  const actionsRef = useRef(useKnowledgeChatStore.getState());
  actionsRef.current = useKnowledgeChatStore.getState();
  const optionsRef = useRef(options);
  optionsRef.current = options;

  // Cleanup listener on unmount
  useEffect(() => {
    return () => {
      if (unlistenRef.current) {
        unlistenRef.current();
        unlistenRef.current = null;
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
        const { id } = await createChatSession();
        sid = id;
        actions.setSessionId(id);
      } catch {
        actions.setError("Failed to create session");
        return;
      }
    }

    actions.addUserMessage(content);

    // Cleanup previous listener
    if (unlistenRef.current) {
      unlistenRef.current();
      unlistenRef.current = null;
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

    // Listen for Tauri events
    const eventName = `chat-event:${sid}`;
    unlistenRef.current = await listen<SSEEvent>(eventName, (event) => {
      processEvent(
        event.payload.type,
        JSON.stringify(event.payload.data),
        wrappedActions,
      );
    });

    try {
      await sendChatMessage(sid, content);
    } catch (err) {
      actionsRef.current.setError(
        err instanceof Error ? err.message : "Connection error",
      );
    }
  }, []);

  const stop = useCallback(() => {
    if (unlistenRef.current) {
      unlistenRef.current();
      unlistenRef.current = null;
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
