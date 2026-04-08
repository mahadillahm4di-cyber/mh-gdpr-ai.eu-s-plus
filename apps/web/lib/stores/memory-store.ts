"use client";

import { create } from "zustand";
import type { Memory } from "@/lib/api";

interface MemoryState {
  memories: Memory[];
  selectedMemory: Memory | null;
  isLoading: boolean;

  // Actions
  setMemories: (m: Memory[]) => void;
  addMemory: (m: Memory) => void;
  removeMemory: (id: string) => void;
  selectMemory: (m: Memory | null) => void;
  setLoading: (l: boolean) => void;
}

export const useMemoryStore = create<MemoryState>((set) => ({
  memories: [],
  selectedMemory: null,
  isLoading: false,

  setMemories: (memories) => set({ memories }),
  addMemory: (m) => set((state) => ({ memories: [m, ...state.memories] })),
  removeMemory: (id) =>
    set((state) => ({
      memories: state.memories.filter((m) => m.id !== id),
      selectedMemory:
        state.selectedMemory?.id === id ? null : state.selectedMemory,
    })),
  selectMemory: (m) => set({ selectedMemory: m }),
  setLoading: (l) => set({ isLoading: l }),
}));
