"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { useChatStore } from "@/lib/stores/chat-store";
import { api } from "@/lib/api";
import type { Memory, ChatMessage as ChatMsg } from "@/lib/api";
import { ProviderSwitch } from "@/components/chat/provider-switch";
import { ChatMessage } from "@/components/chat/chat-message";
import { ChatInput } from "@/components/chat/chat-input";

const MemoryUniverse = dynamic(
  () =>
    import("@/components/space/memory-universe").then(
      (mod) => mod.MemoryUniverse
    ),
  { ssr: false }
);

export default function DashboardPage() {
  const {
    messages,
    provider,
    model,
    isStreaming,
    addMessage,
    appendToLastMessage,
    setStreaming,
  } = useChatStore();

  const [memories, setMemories] = useState<Memory[]>([]);
  const [selectedMemory, setSelectedMemory] = useState<Memory | null>(null);
  const [deleteFlash, setDeleteFlash] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Load memories on mount
  const loadMemories = useCallback(async () => {
    try {
      const data = await api.getMemories();
      setMemories(data.memories || []);
    } catch {
      // silent
    }
  }, []);

  useEffect(() => {
    loadMemories();
  }, [loadMemories]);

  // Auto-scroll chat
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // Detect delete intent and handle it
  const handleDeleteIntent = useCallback(
    async (content: string): Promise<boolean> => {
      const lower = content.toLowerCase();
      const isDelete =
        (lower.includes("supprim") || lower.includes("delete") || lower.includes("efface") || lower.includes("enlev") || lower.includes("retir")) &&
        (lower.includes("souvenir") || lower.includes("memory") || lower.includes("memor") || lower.includes("etoile") || lower.includes("star"));

      if (!isDelete) return false;

      // Search for matching memories
      const keywords = lower
        .replace(/supprime|supprimer|delete|efface|effacer|enlever|enleve|retirer|retire|le|la|les|un|une|du|des|de|mon|ma|mes|ce|cette|souvenir|memory|memoire|etoile|star|sur|qui|parle|about|s'il|te|plait|please/g, "")
        .trim()
        .split(/\s+/)
        .filter((w) => w.length > 2);

      if (keywords.length === 0 && memories.length > 0) {
        // No specific keyword — try to match the last memory
        return false;
      }

      let toDelete: Memory | null = null;
      for (const mem of memories) {
        const summary = mem.summary.toLowerCase();
        if (keywords.some((kw) => summary.includes(kw))) {
          toDelete = mem;
          break;
        }
      }

      if (toDelete) {
        try {
          await api.deleteMemory(toDelete.id);
          setDeleteFlash(toDelete.id);
          setTimeout(() => {
            setMemories((prev) => prev.filter((m) => m.id !== toDelete!.id));
            setDeleteFlash(null);
          }, 600);
          return true;
        } catch {
          return false;
        }
      }
      return false;
    },
    [memories]
  );

  const handleSend = useCallback(
    async (content: string) => {
      // Check for delete intent
      const wasDelete = await handleDeleteIntent(content);

      addMessage({ role: "user", content });
      addMessage({ role: "assistant", content: "" });
      setStreaming(true);

      const systemContext = wasDelete
        ? "The user asked to delete a memory and it was successfully deleted. Confirm the deletion briefly."
        : "";

      const chatMessages: ChatMsg[] = [
        ...(systemContext
          ? [{ role: "system" as const, content: systemContext }]
          : []),
        ...messages.filter((m) => m.content.trim() !== ""),
        { role: "user" as const, content },
      ];

      try {
        await api.chatStream(
          provider,
          { messages: chatMessages, model, stream: true },
          (chunk) => appendToLastMessage(chunk),
          () => {
            setStreaming(false);
            // Reload memories after response (new ones may have been created)
            setTimeout(loadMemories, 500);
          }
        );
      } catch {
        appendToLastMessage(
          "\n\n[Error: Failed to get response. Check your API keys and provider.]"
        );
        setStreaming(false);
      }
    },
    [messages, provider, addMessage, appendToLastMessage, setStreaming, handleDeleteIntent, loadMemories]
  );

  return (
    <div className="flex h-screen bg-black">
      {/* LEFT — Live Memory Space */}
      <div className="relative flex w-[40%] flex-col border-r border-white/10">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
          <Link href="/" className="text-sm font-bold text-white/80 hover:text-white">
            mh-gdpr-ai.eu
          </Link>
          <span className="text-xs text-white/30">
            {memories.length} memories
          </span>
        </div>

        {/* 3D Space */}
        <div className="relative flex-1">
          <MemoryUniverse
            memories={memories}
            onSelectMemory={(m) => setSelectedMemory(m)}
          />

          {/* Delete flash overlay */}
          {deleteFlash && (
            <div className="pointer-events-none absolute inset-0 animate-pulse bg-red-500/10" />
          )}

          {/* Selected memory detail */}
          {selectedMemory && (
            <div className="absolute bottom-4 left-4 right-4 rounded-xl border border-white/10 bg-black/80 p-4 backdrop-blur-md">
              <div className="mb-1 flex items-center justify-between">
                <span className="rounded-full bg-white/10 px-2 py-0.5 text-xs text-white/50">
                  {selectedMemory.theme}
                </span>
                <button
                  onClick={() => setSelectedMemory(null)}
                  className="text-white/30 hover:text-white"
                >
                  x
                </button>
              </div>
              <p className="mt-2 text-sm leading-relaxed text-white/80">
                {selectedMemory.summary.length > 200
                  ? selectedMemory.summary.slice(0, 200) + "..."
                  : selectedMemory.summary}
              </p>
            </div>
          )}
        </div>

        {/* Nav links */}
        <div className="flex gap-2 border-t border-white/10 p-3">
          <Link
            href="/space"
            className="flex-1 rounded-lg border border-white/10 py-2 text-center text-xs text-white/40 hover:bg-white/5 hover:text-white"
          >
            Full Space
          </Link>
          <Link
            href="/settings"
            className="flex-1 rounded-lg border border-white/10 py-2 text-center text-xs text-white/40 hover:bg-white/5 hover:text-white"
          >
            Settings
          </Link>
        </div>
      </div>

      {/* RIGHT — Chat */}
      <div className="flex w-[60%] flex-col">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-white/10 px-6 py-3">
          <ProviderSwitch />
          <Link
            href="/chat"
            className="text-xs text-white/30 hover:text-white"
          >
            Full Chat
          </Link>
        </div>

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
                  Your memories appear live on the left as you chat.
                </p>
                <p className="mt-2 text-xs text-white/20">
                  Say &quot;delete the memory about...&quot; to remove one.
                </p>
              </div>
            </div>
          ) : (
            <div className="mx-auto max-w-2xl">
              {messages.map((msg, i) => (
                <ChatMessage key={i} message={msg} />
              ))}
              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* Input */}
        <div className="border-t border-white/10 px-6 py-4">
          <div className="mx-auto max-w-2xl">
            <ChatInput onSend={handleSend} disabled={isStreaming} />
          </div>
        </div>
      </div>
    </div>
  );
}
