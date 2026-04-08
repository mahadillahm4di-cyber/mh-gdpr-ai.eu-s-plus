"use client";

import { useEffect, useState, useCallback } from "react";
import { api } from "@/lib/api";
import type { Memory } from "@/lib/api";

export function MemoryPanel() {
  const [memories, setMemories] = useState<Memory[]>([]);
  const [loading, setLoading] = useState(true);
  const [flash, setFlash] = useState<string | null>(null);

  const fetchMemories = useCallback(async () => {
    try {
      const data = await api.getMemories();
      setMemories((prev) => {
        // Detect new memories for flash animation
        if (prev.length > 0 && data.memories.length > prev.length) {
          const newMem = data.memories.find(
            (m) => !prev.some((p) => p.id === m.id)
          );
          if (newMem) {
            setFlash(newMem.id);
            setTimeout(() => setFlash(null), 2000);
          }
        }
        return data.memories || [];
      });
    } catch {
      // Silently fail — panel is optional
    } finally {
      setLoading(false);
    }
  }, []);

  // Poll for memory changes every 3 seconds
  useEffect(() => {
    fetchMemories();
    const interval = setInterval(fetchMemories, 3000);
    return () => clearInterval(interval);
  }, [fetchMemories]);

  const importanceColor = (importance: number) => {
    if (importance >= 0.8) return "#F55036";
    if (importance >= 0.5) return "#FBBF24";
    return "#4ADE80";
  };

  return (
    <div className="flex h-full flex-col border-r border-white/10 bg-black/40">
      {/* Header */}
      <div className="border-b border-white/10 px-4 py-3">
        <div className="flex items-center gap-2">
          <span className="text-sm">&#10022;</span>
          <h2 className="text-sm font-semibold text-white/80">Souvenirs</h2>
          <span className="ml-auto rounded-full bg-white/10 px-2 py-0.5 text-xs text-white/40">
            {memories.length}
          </span>
        </div>
        <p className="mt-1 text-xs text-white/30">
          Synchronis&eacute;s entre tous les mod&egrave;les
        </p>
      </div>

      {/* Memory list */}
      <div className="flex-1 overflow-y-auto px-3 py-3">
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-white/20 border-t-white/60" />
          </div>
        ) : memories.length === 0 ? (
          <div className="py-8 text-center">
            <div className="mb-2 text-2xl opacity-30">&#10022;</div>
            <p className="text-xs text-white/30">
              Aucun souvenir pour l&apos;instant.
            </p>
            <p className="mt-1 text-xs text-white/20">
              Dis &quot;souviens-toi que...&quot; pour en cr&eacute;er un.
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {memories.map((mem) => (
              <div
                key={mem.id}
                className={`rounded-lg border p-3 transition-all duration-500 ${
                  flash === mem.id
                    ? "border-green-400/50 bg-green-400/10 shadow-lg shadow-green-400/10"
                    : "border-white/5 bg-white/5 hover:border-white/10 hover:bg-white/8"
                }`}
              >
                {/* Theme + importance */}
                <div className="mb-1.5 flex items-center gap-2">
                  {mem.theme && (
                    <span className="rounded bg-white/10 px-1.5 py-0.5 text-[10px] text-white/40">
                      {mem.theme}
                    </span>
                  )}
                  <div className="ml-auto flex items-center gap-1">
                    <div
                      className="h-1.5 w-1.5 rounded-full"
                      style={{ backgroundColor: importanceColor(mem.importance) }}
                    />
                    <span className="text-[10px] text-white/30">
                      {Math.round(mem.importance * 100)}%
                    </span>
                  </div>
                </div>

                {/* Content */}
                <p className="text-xs leading-relaxed text-white/70">
                  {mem.summary.length > 150
                    ? mem.summary.slice(0, 150) + "..."
                    : mem.summary}
                </p>

                {/* Date */}
                <p className="mt-1.5 text-[10px] text-white/20">
                  {new Date(mem.created_at).toLocaleDateString("fr-FR", {
                    day: "numeric",
                    month: "short",
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
