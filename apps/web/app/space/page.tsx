"use client";

import { useEffect } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { useMemoryStore } from "@/lib/stores/memory-store";
import { api } from "@/lib/api";
import type { Memory } from "@/lib/api";

// Dynamic import — Three.js must be client-side only
const MemoryUniverse = dynamic(
  () =>
    import("@/components/space/memory-universe").then(
      (mod) => mod.MemoryUniverse
    ),
  { ssr: false, loading: () => <LoadingScreen /> }
);

function LoadingScreen() {
  return (
    <div className="flex h-screen items-center justify-center bg-black">
      <div className="text-center">
        <div className="mb-4 animate-pulse text-4xl">✦</div>
        <div className="text-sm text-white/40">Loading your memory space...</div>
      </div>
    </div>
  );
}

export default function SpacePage() {
  const { memories, selectedMemory, setMemories, selectMemory, setLoading } =
    useMemoryStore();

  // Load memories on mount
  useEffect(() => {
    const loadMemories = async () => {
      setLoading(true);
      try {
        const res = await api.getMemories();
        setMemories(res.memories || []);
      } catch {
        // If no auth or no memories, show demo data
        setMemories(generateDemoMemories());
      } finally {
        setLoading(false);
      }
    };
    loadMemories();
  }, [setMemories, setLoading]);

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-black">
      {/* 3D Canvas — Full screen */}
      <MemoryUniverse memories={memories} onSelectMemory={selectMemory} />

      {/* Overlay UI */}
      <div className="pointer-events-none absolute inset-0">
        {/* Top bar */}
        <div className="pointer-events-auto flex items-center justify-between p-4">
          <Link
            href="/"
            className="rounded-lg border border-white/10 bg-black/50 px-4 py-2 text-sm text-white/60 backdrop-blur-sm transition hover:text-white"
          >
            ← Back
          </Link>

          <div className="rounded-lg border border-white/10 bg-black/50 px-4 py-2 text-sm text-white/40 backdrop-blur-sm">
            {memories.length} memories
          </div>

          <Link
            href="/chat"
            className="rounded-lg border border-white/10 bg-black/50 px-4 py-2 text-sm text-white/60 backdrop-blur-sm transition hover:text-white"
          >
            Chat →
          </Link>
        </div>

        {/* Bottom legend */}
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2">
          <div className="pointer-events-auto flex items-center gap-6 rounded-full border border-white/10 bg-black/50 px-6 py-2 text-xs text-white/40 backdrop-blur-sm">
            <span className="flex items-center gap-2">
              <span className="h-2 w-2 rounded-full" style={{ backgroundColor: "#4A90D9" }} />
              OpenAI
            </span>
            <span className="flex items-center gap-2">
              <span className="h-2 w-2 rounded-full" style={{ backgroundColor: "#E87B35" }} />
              Anthropic
            </span>
            <span className="flex items-center gap-2">
              <span className="h-2 w-2 rounded-full" style={{ backgroundColor: "#4ADE80" }} />
              Ollama
            </span>
          </div>
        </div>

        {/* Selected memory panel */}
        {selectedMemory && (
          <div className="pointer-events-auto absolute right-4 top-16 w-80">
            <div className="rounded-xl border border-white/10 bg-black/80 p-4 backdrop-blur-md">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-semibold text-white">
                  {selectedMemory.theme || "Memory"}
                </span>
                <button
                  onClick={() => selectMemory(null)}
                  className="text-white/40 hover:text-white"
                >
                  ✕
                </button>
              </div>
              <p className="mb-3 text-sm text-white/60">{selectedMemory.summary}</p>
              <div className="flex items-center justify-between text-xs text-white/30">
                <span>
                  Importance: {Math.round(selectedMemory.importance * 100)}%
                </span>
                <span>{new Date(selectedMemory.created_at).toLocaleDateString()}</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

/**
 * Generate demo memories for visualization when no real data exists.
 * Positions are spread in 3D space with clustering by theme.
 */
function generateDemoMemories(): Memory[] {
  const themes = [
    { name: "openai", summaries: ["Built a chatbot with GPT-4", "Analyzed data with GPT", "Generated code with GPT"] },
    { name: "anthropic", summaries: ["Discussed project architecture with Claude", "Code review with Claude", "Writing docs with Claude"] },
    { name: "ollama", summaries: ["Tested Llama 3 locally", "Fine-tuned local model", "Ran offline inference"] },
  ];

  const memories: Memory[] = [];
  let id = 0;

  for (const theme of themes) {
    const cx = (Math.random() - 0.5) * 6;
    const cy = (Math.random() - 0.5) * 4;
    const cz = (Math.random() - 0.5) * 4;

    for (const summary of theme.summaries) {
      memories.push({
        id: `demo-${id++}`,
        summary,
        theme: theme.name,
        importance: 0.3 + Math.random() * 0.7,
        position_x: cx + (Math.random() - 0.5) * 2,
        position_y: cy + (Math.random() - 0.5) * 2,
        position_z: cz + (Math.random() - 0.5) * 2,
        created_at: new Date(Date.now() - Math.random() * 30 * 86400000).toISOString(),
      });
    }
  }

  return memories;
}
