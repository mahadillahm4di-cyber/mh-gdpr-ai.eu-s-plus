"use client";

import { useCallback, useState } from "react";
import Link from "next/link";
import { useChatStore } from "@/lib/stores/chat-store";
import { api } from "@/lib/api";
import { ProviderSwitch } from "@/components/chat/provider-switch";
import { ChatMessage } from "@/components/chat/chat-message";
import { ChatInput } from "@/components/chat/chat-input";
import { ConversationSidebar } from "@/components/chat/conversation-sidebar";

export default function ChatPage() {
  const {
    messages,
    provider,
    isStreaming,
    conversationId,
    addMessage,
    appendToLastMessage,
    setStreaming,
  } = useChatStore();

  const [sidebarOpen, setSidebarOpen] = useState(false);

  const handleSend = useCallback(
    async (content: string) => {
      addMessage({ role: "user", content });
      addMessage({ role: "assistant", content: "" });
      setStreaming(true);

      try {
        await api.chatStream(
          provider,
          {
            messages: [...messages, { role: "user" as const, content }],
            stream: true,
          },
          (chunk) => {
            appendToLastMessage(chunk);
          },
          () => {
            setStreaming(false);
          }
        );
      } catch {
        appendToLastMessage(
          "\n\n[Error: Failed to get response. Check your API keys and provider.]"
        );
        setStreaming(false);
      }
    },
    [messages, provider, addMessage, appendToLastMessage, setStreaming]
  );

  return (
    <div className="flex h-screen flex-col">
      {/* Sidebar */}
      <ConversationSidebar
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onSelect={() => setSidebarOpen(false)}
        activeId={conversationId}
      />

      {/* Header */}
      <header className="flex items-center justify-between border-b border-white/10 px-6 py-3">
        <div className="flex items-center gap-4">
          <button
            onClick={() => setSidebarOpen(true)}
            className="text-white/50 hover:text-white"
            title="History"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </button>
          <Link
            href="/"
            className="text-lg font-bold text-white/80 hover:text-white"
          >
            mh-gdpr-ai.eu
          </Link>
          <ProviderSwitch />
        </div>
        <Link
          href="/space"
          className="rounded-lg border border-white/10 px-4 py-2 text-sm text-white/60 transition hover:bg-white/5 hover:text-white"
        >
          Memory Space
        </Link>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-6 py-6">
        {messages.length === 0 ? (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <div className="mb-4 text-4xl">&#10022;</div>
              <h2 className="mb-2 text-xl font-semibold text-white/60">
                Start a conversation
              </h2>
              <p className="text-sm text-white/30">
                Switch providers anytime. Your memory follows.
              </p>
            </div>
          </div>
        ) : (
          <div className="mx-auto max-w-3xl">
            {messages.map((msg, i) => (
              <ChatMessage key={i} message={msg} />
            ))}
          </div>
        )}
      </div>

      {/* Input */}
      <div className="border-t border-white/10 px-6 py-4">
        <div className="mx-auto max-w-3xl">
          <ChatInput onSend={handleSend} disabled={isStreaming} />
        </div>
      </div>
    </div>
  );
}
