<div align="center">

![mh-gdpr-ai.eu S+](docs/assets/logo-banner.png)

# mh-gdpr-ai.eu S+

**Your AI Memory. Your Rules.**

Switch between GPT, Claude, and Llama without losing context. Your memory stays on your machine. Visualize your AI brain in 3D.

<br>

![License](https://img.shields.io/github/license/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus?style=flat-square)
![Stars](https://img.shields.io/github/stars/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus?style=flat-square)
![Watchers](https://img.shields.io/github/watchers/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus?style=flat-square)
![Forks](https://img.shields.io/github/forks/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus?style=flat-square)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)
![Go](https://img.shields.io/badge/backend-Go%201.22-00ADD8?style=flat-square&logo=go)
![Next.js](https://img.shields.io/badge/frontend-Next.js%2015-black?style=flat-square&logo=next.js)
![Three.js](https://img.shields.io/badge/3D-Three.js-black?style=flat-square&logo=three.js)

<br>

[![LinkedIn](https://img.shields.io/badge/LinkedIn-Mahadillah-0A66C2?style=flat-square&logo=linkedin)](https://linkedin.com/in/mahadillah)
[![Email](https://img.shields.io/badge/Email-mahadillahm4di%40proton.me-8B89CC?style=flat-square&logo=protonmail)](mailto:mahadillahm4di@proton.me)
[![GitHub](https://img.shields.io/badge/GitHub-mahadillahm4di--cyber-181717?style=flat-square&logo=github)](https://github.com/mahadillahm4di-cyber)

<br>

[Overview](#-overview) · [Screenshots](#-screenshots) · [Demo](#-demo-videos) · [How it Works](#-workflow) · [Quick Start](#-quick-start) · [API](#api) · [Security](#-security) · [Contributing](#contributing)

</div>

---

## Overview

> **mh-gdpr-ai.eu S+** is a next-generation open-source AI infrastructure protocol. It acts as a universal proxy between you and any AI provider (OpenAI, Anthropic, Ollama). Every conversation is automatically saved locally with AES-256 encryption. When you switch providers, your full context is injected into the new one — the new AI knows everything the previous one knew. Your memory never leaves your machine. Visualize your entire AI brain as stars floating in a 3D spatial dashboard.

### How it works in practice

- You chat with **GPT** → your conversation is saved and encrypted locally
- You switch to **Claude** → the protocol injects your memory → Claude knows everything GPT knew
- You switch to **Llama** (local, free) → same thing, zero context lost
- You open `/space` → your memories float as glowing stars in a 3D void, connected by theme

### Our Vision

**At the infrastructure level:** We are building the missing layer between users and AI providers — a universal memory protocol that makes provider lock-in obsolete. Your data, your rules, your memory.

**At the user level:** We are creating a visual, immersive way to explore your AI brain. Each memory is a star. Each theme is a constellation. Switch providers and watch context flow as light particles between stars.

From serious AI workflows to visual exploration of your own memory, mh-gdpr-ai.eu S+ makes AI truly yours.

---

## Live Demo

> Welcome to our online demo platform! (Coming soon)
>
> **[mh-gdpr-ai.eu](https://mh-gdpr-ai.eu)** — Live demo will be available here after deployment.

---

## Screenshots

| Landing Page | Chat Interface | Provider Switch |
|:---:|:---:|:---:|
| ![Landing](docs/assets/screenshot-landing.png) | ![Chat](docs/assets/screenshot-chat.png) | ![Switch](docs/assets/screenshot-chat-switch.png) |
| **3D Memory Dashboard** | **Memory Detail** | **Login** |
| ![3D](docs/assets/screenshot-space-3d.png) | ![Detail](docs/assets/screenshot-space-detail.png) | ![Login](docs/assets/screenshot-login.png) |

---

## Demo Videos

### 1. Full Protocol Demo — Switch GPT → Claude → Llama without losing context

[![Demo Video](docs/assets/demo-video-cover.png)](https://youtube.com)

*Demo video of mh-gdpr-ai.eu S+*

*Click the image to watch the full demonstration of switching between AI providers with zero context loss.*

### 2. 3D Memory Dashboard — Explore your AI brain in space

[![3D Demo](docs/assets/demo-3d-cover.png)](https://youtube.com)

*3D Dashboard demo of mh-gdpr-ai.eu S+*

*Click the image to watch the full visual exploration of the 3D memory dashboard with stars, connections, and context flows.*

> More demos coming soon: multi-AI collaboration, GDPR plugin, smart routing...

---

## Workflow

1. **Proxy Intercept** — Your message is intercepted by the protocol. The header `X-MH-Provider` tells it which AI to use.
2. **Memory Save** — Every message is encrypted (AES-256-GCM) and saved to your local SQLite database. Automatic.
3. **Switch Detection** — When you change provider, the protocol detects it instantly.
4. **Context Injection** — Your memories and recent conversation are injected as a system prompt into the new provider. It now knows everything.
5. **3D Visualization** — Every memory becomes a star in your 3D dashboard. Color = provider. Size = importance. Lines = shared themes.

---

## Quick Start

Requires: **Docker 20+** and **Git**.

```bash
git clone https://github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus.git
cd mh-gdpr-ai.eu-s-plus
cp .env.example .env
# Add your API keys to .env
docker-compose up --build

# → API:       http://localhost:8080
# → Frontend:  http://localhost:3000
```

---

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
| Smart router (cheapest/fastest) | Coming soon |
| Multi-AI collaboration | Coming soon |
| GDPR plugin (mh-gdpr-ai.eu) | Coming soon |
| SDKs (Python, TypeScript, Go) | Coming soon |

## API

All routes under `/api/v1/` — chat proxy, memories, conversations, providers. Set provider via header: `X-MH-Provider: openai | anthropic | ollama`

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

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) (coming soon).

## License

Apache 2.0 — See [LICENSE](LICENSE).

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus&type=Date)](https://star-history.com/#mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus&Date)

---

<div align="center">

**Made by [Mahadillah](https://github.com/mahadillahm4di-cyber)**

*Your memory belongs to you. Not to OpenAI. Not to Google. Not to anyone.*

</div>
