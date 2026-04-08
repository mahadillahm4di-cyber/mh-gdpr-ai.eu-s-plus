"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Conversation } from "@/lib/api";

interface ConversationSidebarProps {
  isOpen: boolean;
  onClose: () => void;
  onSelect: (conversationId: string) => void;
  activeId: string | null;
}

export function ConversationSidebar({
  isOpen,
  onClose,
  onSelect,
  activeId,
}: ConversationSidebarProps) {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!isOpen) return;

    const load = async () => {
      setIsLoading(true);
      try {
        const res = await api.getConversations();
        setConversations(res.conversations || []);
      } catch {
        setConversations([]);
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, [isOpen]);

  if (!isOpen) return null;

  const providerColor: Record<string, string> = {
    openai: "#4A90D9",
    anthropic: "#E87B35",
    ollama: "#4ADE80",
  };

  return (
    <div className="fixed inset-y-0 left-0 z-40 flex">
      {/* Panel */}
      <div className="w-72 border-r border-white/10 bg-black/95 backdrop-blur-sm">
        <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
          <h2 className="text-sm font-semibold text-white/70">Conversations</h2>
          <button
            onClick={onClose}
            className="text-white/40 hover:text-white"
          >
            ✕
          </button>
        </div>

        <div className="overflow-y-auto p-2" style={{ maxHeight: "calc(100vh - 50px)" }}>
          {isLoading ? (
            <div className="py-8 text-center text-sm text-white/30">Loading...</div>
          ) : conversations.length === 0 ? (
            <div className="py-8 text-center text-sm text-white/30">
              No conversations yet
            </div>
          ) : (
            conversations.map((conv) => (
              <button
                key={conv.id}
                onClick={() => onSelect(conv.id)}
                className={`mb-1 w-full rounded-lg px-3 py-2.5 text-left transition ${
                  activeId === conv.id
                    ? "bg-white/10 text-white"
                    : "text-white/50 hover:bg-white/5 hover:text-white/80"
                }`}
              >
                <div className="flex items-center gap-2">
                  <span
                    className="h-2 w-2 shrink-0 rounded-full"
                    style={{
                      backgroundColor:
                        providerColor[conv.provider] || "#888",
                    }}
                  />
                  <span className="truncate text-sm">
                    {conv.title || "Untitled"}
                  </span>
                </div>
                <div className="mt-1 flex items-center gap-2 text-xs text-white/30">
                  <span>{conv.model}</span>
                  <span>
                    {new Date(conv.updated_at).toLocaleDateString()}
                  </span>
                </div>
              </button>
            ))
          )}
        </div>
      </div>

      {/* Overlay */}
      <div
        className="flex-1 bg-black/50"
        onClick={onClose}
      />
    </div>
  );
}
