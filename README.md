<div align="center">

# ✦ mh-gdpr-ai.eu S+

**Your AI Memory. Your Rules.**

Switch between GPT, Claude, and Llama without losing context.
Your memory stays on your machine. Visualize your AI brain in 3D.

<br>

[Try it](#quick-start) &bull; [How it works](#how-it-works) &bull; [3D Dashboard](#3d-memory-dashboard) &bull; [API](#api) &bull; [Contributing](#contributing)

<br>

![License](https://img.shields.io/badge/license-Apache%202.0-green?style=flat-square)
![Go](https://img.shields.io/badge/backend-Go%201.22-00ADD8?style=flat-square&logo=go)
![Next.js](https://img.shields.io/badge/frontend-Next.js%2015-black?style=flat-square&logo=next.js)
![Three.js](https://img.shields.io/badge/3D-Three.js-black?style=flat-square&logo=three.js)

</div>

---

## The Problem

Every AI provider locks your memory inside their platform.

```
ChatGPT knows your project → switch to Claude → Claude knows NOTHING.
Switch to Llama local → starts from ZERO.

Your memory is imprisoned. You start over. Every. Single. Time.
```

## The Solution

One protocol. All providers. One memory.

```
You → mh-gdpr-ai.eu S+ → GPT / Claude / Llama
                ↓
    Your memory saved locally (encrypted)
                ↓
    Switch provider → context injected automatically
                ↓
    New provider knows EVERYTHING the old one knew.
```

## Quick Start

```bash
# Clone
git clone https://github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus.git
cd mh-gdpr-ai.eu-s-plus

# Copy env
cp .env.example .env
# Add your API keys to .env

# Run with Docker
docker-compose up

# → API:       http://localhost:8080
# → Frontend:  http://localhost:3000
```

## How It Works

```
┌──────────────────────────────────────────┐
│         SOVEREIGN AI PROTOCOL            │
│                                          │
│  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │  PROXY   │→ │  MEMORY  │→ │INJECTOR│ │
│  │          │  │ (SQLite) │  │        │ │
│  │ Intercept│  │ Encrypted│  │ Switch │ │
│  │ all calls│  │ AES-256  │  │ = auto │ │
│  └──────────┘  └──────────┘  └────────┘ │
│                                          │
└─────────────────┬────────────────────────┘
                  │
     ┌────────────┼────────────┐
     ▼            ▼            ▼
  ┌──────┐   ┌────────┐   ┌───────┐
  │OpenAI│   │Anthropic│  │Ollama │
  │ GPT  │   │ Claude  │  │ Llama │
  │(cloud)│  │ (cloud) │  │(local)│
  └──────┘   └────────┘   └───────┘
```

## 3D Memory Dashboard

Your memories visualized as stars in space.

- Each star = one memory
- Color = provider (blue=GPT, orange=Claude, green=Llama)
- Connected stars = related topics
- Click a star = see the full conversation
- Switch provider = watch context flow as light particles

Visit `/space` to explore your AI brain.

## Features

| Feature | Status |
|---------|--------|
| Proxy (OpenAI, Anthropic, Ollama) | ✅ |
| Local memory (SQLite, encrypted) | ✅ |
| Context injection on provider switch | ✅ |
| Streaming responses (SSE) | ✅ |
| 3D spatial dashboard | ✅ |
| Chat interface with provider switch | ✅ |
| JWT auth + security headers | ✅ |
| AES-256 encryption at rest | ✅ |
| Rate limiting | ✅ |
| Docker support | ✅ |
| Smart router (cheapest/fastest) | 🔜 |
| Multi-AI collaboration | 🔜 |
| GDPR plugin (mh-gdpr-ai.eu) | 🔜 |
| SDKs (Python, TypeScript, Go) | 🔜 |

## Security

- All memory encrypted at rest (AES-256-GCM)
- All communication over HTTPS/TLS
- JWT auth with short-lived tokens (15min)
- Security headers on every response
- Rate limiting on all endpoints
- CORS with explicit origins only
- SQL injection protection (parameterized queries)
- No secrets in code (env vars only)
- Docker: non-root, read-only filesystem

See [SECURITY.md](SECURITY.md) for our full security policy.

## API

All routes prefixed with `/api/v1/`.

```
POST /chat/completions     → Proxy to AI provider (streaming)
GET  /memories             → List your memories
GET  /memories/search?q=   → Search memories
GET  /conversations        → List conversations
GET  /providers            → Available providers
```

Set provider via header: `X-MH-Provider: openai | anthropic | ollama`

## Tech Stack

- **Backend**: Go 1.22 + Gin
- **Frontend**: Next.js 15 + TypeScript + Tailwind CSS
- **3D**: React Three Fiber + drei + postprocessing
- **Database**: SQLite (local) — zero config
- **Security**: AES-256-GCM, JWT, bcrypt, security headers

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) (coming soon).

## License

Apache 2.0 — See [LICENSE](LICENSE).

---

<div align="center">

**Made by [Mahadillah](https://github.com/mahadillahm4di-cyber)**

*Your memory belongs to you. Not to OpenAI. Not to Google. Not to anyone.*

</div>
