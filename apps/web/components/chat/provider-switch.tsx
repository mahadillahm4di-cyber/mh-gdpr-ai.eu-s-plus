"use client";

import { useChatStore } from "@/lib/stores/chat-store";
import type { Provider } from "@/lib/api";

interface ModelOption {
  id: string;
  label: string;
}

const providers: { id: Provider; label: string; color: string; models: ModelOption[] }[] = [
  {
    id: "groq",
    label: "Groq",
    color: "#F55036",
    models: [
      { id: "llama-3.3-70b-versatile", label: "Llama 3.3 70B" },
      { id: "openai/gpt-oss-120b", label: "GPT-OSS 120B" },
      { id: "meta-llama/llama-4-scout-17b-16e-instruct", label: "Llama 4 Scout" },
      { id: "qwen/qwen3-32b", label: "Qwen 3 32B" },
      { id: "moonshotai/kimi-k2-instruct", label: "Kimi K2" },
      { id: "llama-3.1-8b-instant", label: "Llama 3.1 8B (rapide)" },
    ],
  },
  {
    id: "openai",
    label: "GPT",
    color: "#4A90D9",
    models: [
      { id: "gpt-4o-mini", label: "GPT-4o Mini" },
    ],
  },
  {
    id: "anthropic",
    label: "Claude",
    color: "#E87B35",
    models: [
      { id: "claude-3-haiku-20240307", label: "Claude 3 Haiku" },
    ],
  },
  {
    id: "ollama",
    label: "Llama (local)",
    color: "#4ADE80",
    models: [
      { id: "tinyllama", label: "TinyLlama" },
    ],
  },
];

export function ProviderSwitch() {
  const { provider, model, setProvider, setModel, isStreaming } = useChatStore();

  const activeProvider = providers.find((p) => p.id === provider) || providers[0];

  const handleProviderChange = (p: Provider) => {
    setProvider(p);
    const newProvider = providers.find((pr) => pr.id === p);
    if (newProvider) {
      setModel(newProvider.models[0].id);
    }
  };

  return (
    <div className="flex items-center gap-2">
      {providers.map((p) => (
        <button
          key={p.id}
          onClick={() => handleProviderChange(p.id)}
          disabled={isStreaming}
          className={`rounded-lg px-3 py-1.5 text-sm font-medium transition-all ${
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
            className="mr-1.5 inline-block h-2 w-2 rounded-full"
            style={{ backgroundColor: p.color }}
          />
          {p.label}
        </button>
      ))}

      {/* Model selector — only show if provider has multiple models */}
      {activeProvider.models.length > 1 && (
        <select
          value={model}
          onChange={(e) => setModel(e.target.value)}
          disabled={isStreaming}
          className="ml-1 rounded-lg border border-white/10 bg-black px-2 py-1.5 text-xs text-white/60 outline-none transition hover:border-white/20 focus:border-white/30 disabled:cursor-not-allowed disabled:opacity-50"
          style={{ maxWidth: "180px" }}
        >
          {provider === "groq" ? (
            <>
              <optgroup label="Puissant" className="bg-black text-white">
                {activeProvider.models.filter((m) => m.id !== "llama-3.1-8b-instant").map((m) => (
                  <option key={m.id} value={m.id} className="bg-black text-white">
                    {m.label}
                  </option>
                ))}
              </optgroup>
              <optgroup label="Rapide" className="bg-black text-white">
                {activeProvider.models.filter((m) => m.id === "llama-3.1-8b-instant").map((m) => (
                  <option key={m.id} value={m.id} className="bg-black text-white">
                    {m.label}
                  </option>
                ))}
              </optgroup>
            </>
          ) : (
            activeProvider.models.map((m) => (
              <option key={m.id} value={m.id} className="bg-black text-white">
                {m.label}
              </option>
            ))
          )}
        </select>
      )}
    </div>
  );
}
