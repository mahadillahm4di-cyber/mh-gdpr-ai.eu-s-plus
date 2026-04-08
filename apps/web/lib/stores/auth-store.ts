"use client";

import { create } from "zustand";
import { api } from "@/lib/api";

interface AuthState {
  isAuthenticated: boolean;
  userId: string | null;
  isLoading: boolean;
  error: string | null;

  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  userId: null,
  isLoading: false,
  error: null,

  login: async (email: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const res = await api.login(email, password);
      api.setToken(res.access_token);
      set({
        isAuthenticated: true,
        userId: res.user_id,
        isLoading: false,
      });
    } catch (err) {
      set({
        isLoading: false,
        error: err instanceof Error ? err.message : "Login failed",
      });
    }
  },

  register: async (email: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const res = await api.register(email, password);
      api.setToken(res.access_token);
      set({
        isAuthenticated: true,
        userId: res.user_id,
        isLoading: false,
      });
    } catch (err) {
      set({
        isLoading: false,
        error: err instanceof Error ? err.message : "Registration failed",
      });
    }
  },

  logout: () => {
    api.setToken("");
    set({ isAuthenticated: false, userId: null, error: null });
  },

  clearError: () => set({ error: null }),
}));
