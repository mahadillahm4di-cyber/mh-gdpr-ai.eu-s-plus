"use client";

import { create } from "zustand";
import type { Provider, ChatMessage } from "@/lib/api";

interface ChatState {
  messages: ChatMessage[];
  provider: Provider;
  isStreaming: boolean;
  conversationId: string | null;

  // Actions
  addMessage: (msg: ChatMessage) => void;
  appendToLastMessage: (content: string) => void;
  setProvider: (p: Provider) => void;
  setStreaming: (s: boolean) => void;
  setConversationId: (id: string) => void;
  clearMessages: () => void;
}

export const useChatStore = create<ChatState>((set) => ({
  messages: [],
  provider: "openai",
  isStreaming: false,
  conversationId: null,

  addMessage: (msg) =>
    set((state) => ({ messages: [...state.messages, msg] })),

  appendToLastMessage: (content) =>
    set((state) => {
      const msgs = [...state.messages];
      if (msgs.length > 0 && msgs[msgs.length - 1].role === "assistant") {
        msgs[msgs.length - 1] = {
          ...msgs[msgs.length - 1],
          content: msgs[msgs.length - 1].content + content,
        };
      }
      return { messages: msgs };
    }),

  setProvider: (p) => set({ provider: p }),
  setStreaming: (s) => set({ isStreaming: s }),
  setConversationId: (id) => set({ conversationId: id }),
  clearMessages: () => set({ messages: [], conversationId: null }),
}));
