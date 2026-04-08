"use client";

import { useChatStore } from "@/lib/stores/chat-store";
import type { Provider } from "@/lib/api";

const providers: { id: Provider; label: string; color: string }[] = [
  { id: "openai", label: "GPT", color: "#4A90D9" },
  { id: "anthropic", label: "Claude", color: "#E87B35" },
  { id: "ollama", label: "Llama", color: "#4ADE80" },
];

export function ProviderSwitch() {
  const { provider, setProvider, isStreaming } = useChatStore();

  return (
    <div className="flex items-center gap-2">
      {providers.map((p) => (
        <button
          key={p.id}
          onClick={() => setProvider(p.id)}
          disabled={isStreaming}
          className={`rounded-lg px-4 py-2 text-sm font-medium transition-all ${
            provider === p.id
              ? "text-white shadow-lg"
              : "text-white/40 hover:text-white/70"
          } disabled:cursor-not-allowed disabled:opacity-50`}
          style={{
            backgroundColor:
              provider === p.id ? p.color + "20" : "transparent",
            borderColor: provider === p.id ? p.color : "transparent",
            borderWidth: "1px",
            boxShadow:
              provider === p.id ? `0 0 20px ${p.color}30` : "none",
          }}
        >
          <span
            className="mr-2 inline-block h-2 w-2 rounded-full"
            style={{ backgroundColor: p.color }}
          />
          {p.label}
        </button>
      ))}
    </div>
  );
}
