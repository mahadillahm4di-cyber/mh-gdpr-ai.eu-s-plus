"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";

interface SettingsState {
  openaiKey: string;
  anthropicKey: string;
  groqKey: string;
  openaiConfigured: boolean;
  anthropicConfigured: boolean;
  groqConfigured: boolean;
  openaiMasked: string;
  anthropicMasked: string;
  groqMasked: string;
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<SettingsState>({
    openaiKey: "",
    anthropicKey: "",
    groqKey: "",
    openaiConfigured: false,
    anthropicConfigured: false,
    groqConfigured: false,
    openaiMasked: "",
    anthropicMasked: "",
    groqMasked: "",
  });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    api.getSettings().then((data) => {
      setSettings((s) => ({
        ...s,
        openaiConfigured: data.openai_configured,
        anthropicConfigured: data.anthropic_configured,
        groqConfigured: data.groq_configured,
        openaiMasked: data.openai_api_key,
        anthropicMasked: data.anthropic_api_key,
        groqMasked: data.groq_api_key,
      }));
    }).catch(() => {});
  }, []);

  const handleSave = useCallback(async () => {
    setSaving(true);
    setMessage("");
    try {
      await api.saveSettings({
        openai_api_key: settings.openaiKey,
        anthropic_api_key: settings.anthropicKey,
        groq_api_key: settings.groqKey,
      });
      setMessage("Saved!");
      setSettings((s) => ({
        ...s,
        openaiKey: "",
        anthropicKey: "",
        groqKey: "",
        openaiConfigured: s.openaiKey !== "" || s.openaiConfigured,
        anthropicConfigured: s.anthropicKey !== "" || s.anthropicConfigured,
        groqConfigured: s.groqKey !== "" || s.groqConfigured,
        openaiMasked: s.openaiKey ? "sk-..." + s.openaiKey.slice(-4) : s.openaiMasked,
        anthropicMasked: s.anthropicKey ? "sk-..." + s.anthropicKey.slice(-4) : s.anthropicMasked,
        groqMasked: s.groqKey ? "gsk-..." + s.groqKey.slice(-4) : s.groqMasked,
      }));
    } catch {
      setMessage("Error saving settings");
    }
    setSaving(false);
  }, [settings]);

  return (
    <div className="flex min-h-screen flex-col bg-black text-white">
      {/* Header */}
      <header className="flex items-center justify-between border-b border-white/10 px-6 py-3">
        <Link href="/" className="text-lg font-bold text-white/80 hover:text-white">
          mh-gdpr-ai.eu
        </Link>
        <div className="flex gap-3">
          <Link
            href="/chat"
            className="rounded-lg border border-white/10 px-4 py-2 text-sm text-white/60 transition hover:bg-white/5 hover:text-white"
          >
            Chat
          </Link>
          <Link
            href="/space"
            className="rounded-lg border border-white/10 px-4 py-2 text-sm text-white/60 transition hover:bg-white/5 hover:text-white"
          >
            Memory Space
          </Link>
        </div>
      </header>

      {/* Content */}
      <div className="mx-auto w-full max-w-lg px-6 py-12">
        <h1 className="mb-2 text-2xl font-bold">Settings</h1>
        <p className="mb-8 text-sm text-white/40">
          Add your API keys to use GPT and Claude. Keys are encrypted and stored locally.
        </p>

        {/* Groq */}
        <div className="mb-6 rounded-xl border border-white/10 bg-white/5 p-4">
          <div className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full" style={{ backgroundColor: "#F55036" }} />
            <span className="font-medium">Groq (Llama)</span>
            <span className="ml-auto rounded-full bg-red-400/10 px-3 py-1 text-xs text-red-400">
              Free
            </span>
            {settings.groqConfigured && (
              <span className="rounded-full bg-red-400/10 px-3 py-1 text-xs text-red-400">
                {settings.groqMasked}
              </span>
            )}
          </div>
          <input
            type="password"
            placeholder="gsk_..."
            value={settings.groqKey}
            onChange={(e) => setSettings((s) => ({ ...s, groqKey: e.target.value }))}
            className="mt-3 w-full rounded-lg border border-white/10 bg-black px-4 py-2.5 text-sm text-white placeholder-white/20 outline-none focus:border-red-500/50"
          />
          <p className="mt-2 text-xs text-white/30">
            Free API key at console.groq.com — Recommended for getting started.
          </p>
        </div>

        {/* OpenAI */}
        <div className="mb-6 rounded-xl border border-white/10 bg-white/5 p-4">
          <div className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full" style={{ backgroundColor: "#4A90D9" }} />
            <span className="font-medium">OpenAI (GPT)</span>
            {settings.openaiConfigured && (
              <span className="ml-auto rounded-full bg-blue-400/10 px-3 py-1 text-xs text-blue-400">
                {settings.openaiMasked}
              </span>
            )}
          </div>
          <input
            type="password"
            placeholder="sk-..."
            value={settings.openaiKey}
            onChange={(e) => setSettings((s) => ({ ...s, openaiKey: e.target.value }))}
            className="mt-3 w-full rounded-lg border border-white/10 bg-black px-4 py-2.5 text-sm text-white placeholder-white/20 outline-none focus:border-blue-500/50"
          />
          <p className="mt-2 text-xs text-white/30">
            Get your key at platform.openai.com/api-keys
          </p>
        </div>

        {/* Anthropic */}
        <div className="mb-8 rounded-xl border border-white/10 bg-white/5 p-4">
          <div className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full" style={{ backgroundColor: "#E87B35" }} />
            <span className="font-medium">Anthropic (Claude)</span>
            {settings.anthropicConfigured && (
              <span className="ml-auto rounded-full bg-orange-400/10 px-3 py-1 text-xs text-orange-400">
                {settings.anthropicMasked}
              </span>
            )}
          </div>
          <input
            type="password"
            placeholder="sk-ant-..."
            value={settings.anthropicKey}
            onChange={(e) => setSettings((s) => ({ ...s, anthropicKey: e.target.value }))}
            className="mt-3 w-full rounded-lg border border-white/10 bg-black px-4 py-2.5 text-sm text-white placeholder-white/20 outline-none focus:border-orange-500/50"
          />
          <p className="mt-2 text-xs text-white/30">
            Get your key at console.anthropic.com/settings/keys
          </p>
        </div>

        {/* Save */}
        <button
          onClick={handleSave}
          disabled={saving || (settings.openaiKey === "" && settings.anthropicKey === "" && settings.groqKey === "")}
          className="w-full rounded-lg bg-white/10 px-6 py-3 font-medium text-white transition hover:bg-white/20 disabled:cursor-not-allowed disabled:opacity-30"
        >
          {saving ? "Saving..." : "Save API Keys"}
        </button>

        {message && (
          <p className={`mt-3 text-center text-sm ${message === "Saved!" ? "text-green-400" : "text-red-400"}`}>
            {message}
          </p>
        )}

        <p className="mt-6 text-center text-xs text-white/20">
          Your keys are encrypted with AES-256 and never leave your machine.
        </p>
      </div>
    </div>
  );
}
