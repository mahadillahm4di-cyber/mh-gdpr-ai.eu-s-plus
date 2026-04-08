"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { useChatStore } from "@/lib/stores/chat-store";
import { api } from "@/lib/api";
import type { Memory } from "@/lib/api";
import { ProviderSwitch } from "@/components/chat/provider-switch";
import { ChatMessage } from "@/components/chat/chat-message";
import { ChatInput } from "@/components/chat/chat-input";
import { ConversationSidebar } from "@/components/chat/conversation-sidebar";

const MemoryUniverse = dynamic(
  () =>
    import("@/components/space/memory-universe").then(
      (mod) => mod.MemoryUniverse
    ),
  { ssr: false }
);

export default function ChatPage() {
  const {
    messages,
    provider,
    isStreaming,
    conversationId,
    splitView,
    addMessage,
    appendToLastMessage,
    setStreaming,
    toggleSplitView,
  } = useChatStore();

  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [showApiBanner, setShowApiBanner] = useState(false);
  const [bannerDismissed, setBannerDismissed] = useState(false);
  const [memories, setMemories] = useState<Memory[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Check if user has their own Groq key — if not, show banner
  useEffect(() => {
    api.getSettings().then((data) => {
      if (!data.groq_configured && !bannerDismissed) {
        setShowApiBanner(true);
      }
    }).catch(() => {});
  }, [bannerDismissed]);

  // Load memories for the 3D stars panel
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
    const interval = setInterval(loadMemories, 5000);
    return () => clearInterval(interval);
  }, [loadMemories]);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

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
            setTimeout(loadMemories, 500);
          }
        );
      } catch {
        appendToLastMessage(
          "\n\n[Erreur : impossible d'obtenir une r\u00e9ponse. V\u00e9rifiez votre cl\u00e9 API et le provider.]"
        );
        setStreaming(false);
      }
    },
    [messages, provider, addMessage, appendToLastMessage, setStreaming]
  );

  return (
    <div className="flex h-screen flex-col bg-black text-white">
      {/* Sidebar */}
      <ConversationSidebar
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onSelect={() => setSidebarOpen(false)}
        activeId={conversationId}
      />

      {/* Header */}
      <header className="flex items-center justify-between border-b border-white/10 px-4 py-2.5">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setSidebarOpen(true)}
            className="text-white/50 hover:text-white"
            title="Historique"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </button>
          <Link href="/" className="text-base font-bold text-white/80 hover:text-white">
            mh-gdpr-ai.eu
          </Link>
          <ProviderSwitch />
        </div>
        <div className="flex items-center gap-2">
          {/* Split view toggle */}
          <button
            onClick={toggleSplitView}
            className={`rounded-lg border px-3 py-1.5 text-xs font-medium transition ${
              splitView
                ? "border-purple-500/30 bg-purple-500/10 text-purple-400"
                : "border-white/10 text-white/40 hover:bg-white/5 hover:text-white/60"
            }`}
            title={splitView ? "Mode chat complet" : "Mode souvenirs + chat"}
          >
            <span className="flex items-center gap-1.5">
              {splitView ? (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="3" y="3" width="18" height="18" rx="2" />
                  </svg>
                  Chat
                </>
              ) : (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="3" y="3" width="18" height="18" rx="2" />
                    <line x1="9" y1="3" x2="9" y2="21" />
                  </svg>
                  Souvenirs
                </>
              )}
            </span>
          </button>
          <Link
            href="/settings"
            className="rounded-lg border border-white/10 px-3 py-1.5 text-xs text-white/40 transition hover:bg-white/5 hover:text-white/60"
          >
            Settings
          </Link>
          <Link
            href="/space"
            className="rounded-lg border border-white/10 px-3 py-1.5 text-xs text-white/40 transition hover:bg-white/5 hover:text-white/60"
          >
            Memory Space
          </Link>
        </div>
      </header>

      {/* API Key Banner */}
      {showApiBanner && (
        <div className="flex items-center justify-between border-b border-amber-500/10 bg-amber-500/5 px-4 py-2">
          <p className="text-xs text-amber-300/80">
            <span className="font-medium">Astuce :</span> Obtenez votre propre cl&eacute; API Groq gratuite sur{" "}
            <a
              href="https://console.groq.com"
              target="_blank"
              rel="noopener noreferrer"
              className="font-medium underline hover:text-amber-200"
            >
              console.groq.com
            </a>
            {" "}puis collez-la dans{" "}
            <Link href="/settings" className="font-medium underline hover:text-amber-200">
              Settings
            </Link>
            . C&apos;est gratuit et rapide !
            <span className="ml-2 text-amber-400/60">
              Notez votre cl&eacute; dans un endroit s&ucirc;r.
            </span>
          </p>
          <button
            onClick={() => { setShowApiBanner(false); setBannerDismissed(true); }}
            className="ml-3 text-amber-400/40 hover:text-amber-300"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
      )}

      {/* Main content: split or full */}
      <div className="flex flex-1 overflow-hidden">
        {/* Memory Stars 3D (left) */}
        {splitView && (
          <div className="relative w-[25%] flex-shrink-0 border-r border-white/10">
            <MemoryUniverse memories={memories} onSelectMemory={() => {}} />
            <div className="pointer-events-none absolute bottom-3 left-0 right-0 text-center text-xs text-white/20">
              {memories.length} souvenirs
            </div>
          </div>
        )}

        {/* Memory list (middle column) */}
        {splitView && (
          <div className="flex w-[22%] flex-shrink-0 flex-col border-r border-white/10">
            <div className="border-b border-white/10 px-3 py-2.5">
              <h3 className="text-xs font-semibold text-white/60">Souvenirs</h3>
              <p className="mt-0.5 text-[10px] text-white/25">Copiez un texte et collez dans le chat pour le supprimer</p>
            </div>
            <div className="flex-1 overflow-y-auto px-2 py-2">
              {memories.length === 0 ? (
                <p className="py-8 text-center text-xs text-white/20">Aucun souvenir</p>
              ) : (
                <div className="space-y-1.5">
                  {memories.map((mem) => (
                    <div
                      key={mem.id}
                      className="group cursor-pointer rounded-lg border border-white/5 bg-white/[0.03] p-2.5 transition hover:border-white/10 hover:bg-white/[0.06]"
                      onClick={() => {
                        navigator.clipboard.writeText(mem.summary.slice(0, 80));
                      }}
                      title="Cliquez pour copier"
                    >
                      <div className="mb-1 flex items-center gap-1.5">
                        {mem.theme && (
                          <span className="rounded bg-white/10 px-1.5 py-0.5 text-[9px] text-white/40">
                            {mem.theme}
                          </span>
                        )}
                        <span className="ml-auto text-[9px] text-white/20">
                          {new Date(mem.created_at).toLocaleDateString("fr-FR", {
                            day: "numeric",
                            month: "short",
                          })}
                        </span>
                      </div>
                      <p className="select-all text-[11px] leading-relaxed text-white/60">
                        {mem.summary}
                      </p>
                      <p className="mt-1 text-[9px] text-white/15 opacity-0 transition group-hover:opacity-100">
                        cliquer pour copier
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Chat area (right or full) */}
        <div className="flex flex-1 flex-col">
          {/* Messages */}
          <div className="flex-1 overflow-y-auto px-6 py-6">
            {messages.length === 0 ? (
              <div className="flex h-full items-center justify-center">
                <div className="text-center">
                  <div className="mb-4 text-5xl opacity-20">&#10022;</div>
                  <h2 className="mb-2 text-lg font-semibold text-white/50">
                    Commencez une conversation
                  </h2>
                  <p className="text-sm text-white/25">
                    Changez de mod&egrave;le IA &agrave; tout moment. Vos souvenirs suivent.
                  </p>
                </div>
              </div>
            ) : (
              <div className="mx-auto max-w-3xl">
                {messages.map((msg, i) => (
                  <ChatMessage key={i} message={msg} />
                ))}
                <div ref={messagesEndRef} />
              </div>
            )}
          </div>

          {/* Input */}
          <div className="border-t border-white/10 px-6 py-3">
            <div className="mx-auto max-w-3xl">
              <ChatInput onSend={handleSend} disabled={isStreaming} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
