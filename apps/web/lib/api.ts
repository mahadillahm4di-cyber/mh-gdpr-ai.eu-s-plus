/**
 * mh-gdpr-ai.eu S+ — API Client
 *
 * SECURITY: Tokens are stored in httpOnly cookies (set by the server).
 * This client never touches localStorage for auth tokens.
 * All requests include credentials for cookie-based auth.
 */

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

type Provider = "openai" | "anthropic" | "ollama";

interface ChatMessage {
  role: "system" | "user" | "assistant";
  content: string;
}

interface ChatRequest {
  messages: ChatMessage[];
  model?: string;
  stream?: boolean;
}

interface Memory {
  id: string;
  summary: string;
  theme: string;
  importance: number;
  position_x: number;
  position_y: number;
  position_z: number;
  created_at: string;
}

interface Conversation {
  id: string;
  title: string;
  provider: string;
  model: string;
  created_at: string;
  updated_at: string;
}

class MHGatewayAPI {
  private token: string | null = null;

  setToken(token: string) {
    this.token = token;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
      credentials: "include",
    });

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: "Request failed" }));
      throw new Error(error.error || `HTTP ${res.status}`);
    }

    return res.json();
  }

  // ── Auth ──

  async register(email: string, password: string) {
    return this.request<{ access_token: string; refresh_token: string; user_id: string }>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  }

  async login(email: string, password: string) {
    return this.request<{ access_token: string; refresh_token: string; user_id: string }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  }

  async refresh(refreshToken: string) {
    return this.request<{ access_token: string; refresh_token: string }>("/api/v1/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  }

  // ── Chat ──

  async chat(provider: Provider, req: ChatRequest) {
    return this.request<{
      id: string;
      content: string;
      provider: string;
      model: string;
      tokens_in: number;
      tokens_out: number;
    }>("/api/v1/chat/completions", {
      method: "POST",
      headers: { "X-MH-Provider": provider },
      body: JSON.stringify(req),
    });
  }

  async chatStream(
    provider: Provider,
    req: ChatRequest,
    onChunk: (content: string) => void,
    onDone: () => void
  ) {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-MH-Provider": provider,
    };

    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const res = await fetch(`${API_BASE}/api/v1/chat/completions`, {
      method: "POST",
      headers,
      body: JSON.stringify({ ...req, stream: true }),
      credentials: "include",
    });

    if (!res.ok) {
      throw new Error(`Stream failed: HTTP ${res.status}`);
    }

    const reader = res.body?.getReader();
    if (!reader) throw new Error("No response body");

    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";

      for (const line of lines) {
        if (line.startsWith("data: ")) {
          const data = line.slice(6);
          try {
            const chunk = JSON.parse(data);
            if (chunk.done) {
              onDone();
              return;
            }
            if (chunk.content) {
              onChunk(chunk.content);
            }
          } catch {
            // Skip malformed chunks
          }
        }
      }
    }

    onDone();
  }

  // ── Memories ──

  async getMemories() {
    return this.request<{ memories: Memory[] }>("/api/v1/memories");
  }

  async searchMemories(query: string) {
    return this.request<{ memories: Memory[] }>(
      `/api/v1/memories/search?q=${encodeURIComponent(query)}`
    );
  }

  async deleteMemory(id: string) {
    return this.request<{ deleted: boolean }>(`/api/v1/memories/${id}`, {
      method: "DELETE",
    });
  }

  // ── Conversations ──

  async getConversations(limit = 50, offset = 0) {
    return this.request<{ conversations: Conversation[] }>(
      `/api/v1/conversations?limit=${limit}&offset=${offset}`
    );
  }

  async getConversation(id: string) {
    return this.request<{ conversation: Conversation; messages: ChatMessage[] }>(
      `/api/v1/conversations/${id}`
    );
  }

  async deleteConversation(id: string) {
    return this.request<{ deleted: boolean }>(`/api/v1/conversations/${id}`, {
      method: "DELETE",
    });
  }

  // ── Providers ──

  async getProviders() {
    return this.request<{
      providers: { name: string; default_model: string; healthy: boolean }[];
    }>("/api/v1/providers");
  }
}

export const api = new MHGatewayAPI();
export type { Provider, ChatMessage, Memory, Conversation };
