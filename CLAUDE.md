# mh-gdpr-ai.eu S+ — CLAUDE.md

## IDENTITÉ DU PROJET

- **Nom** : mh-gdpr-ai.eu S+
- **Repo** : mh-gdpr-ai.eu-s-plus
- **Description** : Protocole open-source qui unifie la mémoire IA entre tous les providers
  (OpenAI, Anthropic, Ollama) avec un dashboard spatial 3D pour visualiser sa mémoire.
  L'utilisateur switch d'IA sans jamais perdre son contexte. Tout reste en local.
- **Auteur** : Mahadillah
- **License** : Apache 2.0

---

## ARCHITECTURE

Monorepo avec 2 applications + packages.

```
mh-gdpr-ai.eu-s-plus/
├── apps/api/        ← Backend Go (protocole, proxy, mémoire, routeur)
├── apps/web/        ← Frontend Next.js (site, chat, dashboard 3D)
├── packages/        ← Plugins et SDKs
├── CLAUDE.md        ← CE FICHIER — règles du projet
└── CLAUDE_GUIDE.md  ← Guide étape par étape (0 → MVP → complet)
```

### Backend — Go (apps/api/)
- **Langage** : Go 1.22+
- **Framework** : Gin
- **Base de données** : SQLite (local), PostgreSQL (cloud optionnel)
- **Port** : 8080
- **Rôle** : proxy IA, stockage mémoire, routeur intelligent, injecteur de contexte
- **Toutes les routes** : préfixées par `/api/v1/`

### Frontend — TypeScript (apps/web/)
- **Framework** : Next.js 15 (App Router)
- **3D** : React Three Fiber + @react-three/drei + @react-three/postprocessing
- **Style** : Tailwind CSS 4 + shadcn/ui
- **State** : Zustand
- **Validation** : Zod
- **Port** : 3000

### Communication Backend ↔ Frontend
- **REST** : pour mémoire, routeur, settings, auth
- **WebSocket** : pour le streaming des réponses IA en temps réel
- **Format** : JSON uniquement

---

## PROVIDERS IA SUPPORTÉS (MVP)

| Provider | Endpoint | Type | Adaptateur |
|----------|----------|------|------------|
| OpenAI | api.openai.com | Cloud (payant) | `internal/proxy/openai.go` |
| Anthropic | api.anthropic.com | Cloud (payant) | `internal/proxy/anthropic.go` |
| Ollama | localhost:11434 | Local (gratuit) | `internal/proxy/ollama.go` |

L'utilisateur choisit le provider via le header `X-MH-Provider`.
Le proxy convertit TOUTES les requêtes en format unifié interne, puis adapte au provider cible.

---

## CONVENTIONS GO (apps/api/)

### Structure
```
apps/api/
├── cmd/server/main.go          ← SEUL entrypoint, lance le serveur
├── internal/                   ← Code privé, non importable
│   ├── proxy/                  ← Intercepte et forward les appels IA
│   ├── memory/                 ← Stockage mémoire SQLite
│   ├── router/                 ← Choisit le meilleur provider
│   ├── injector/               ← Injecte le contexte au switch
│   ├── auth/                   ← Authentification JWT
│   ├── config/                 ← Chargement config depuis env
│   └── middleware/             ← CORS, rate limit, security headers, logging
├── go.mod
└── Dockerfile
```

### Règles
- **Nommage** : `camelCase` variables locales, `PascalCase` exports
- **Erreurs** : TOUJOURS retourner `error`, JAMAIS `panic` en production
- **Logs** : `slog` (standard library), JAMAIS `fmt.Println`
- **Tests** : `*_test.go` à côté de chaque fichier, `go test ./...`
- **Pas d'ORM** : requêtes SQL directes via `database/sql` ou `sqlc`
- **Context** : passer `context.Context` en premier argument de toute fonction async
- **Goroutines** : toujours avec `context` et timeout pour éviter les leaks

---

## CONVENTIONS TYPESCRIPT (apps/web/)

### Structure
```
apps/web/
├── app/                        ← Next.js App Router pages
│   ├── layout.tsx              ← Layout principal
│   ├── page.tsx                ← Landing page
│   ├── chat/page.tsx           ← Interface chat
│   └── space/page.tsx          ← Dashboard 3D spatial
├── components/
│   ├── ui/                     ← shadcn/ui
│   ├── chat/                   ← Composants chat
│   └── space/                  ← Composants 3D
├── lib/
│   ├── api.ts                  ← Client API vers le backend
│   ├── stores/                 ← Zustand stores
│   ├── schemas.ts              ← Schémas Zod
│   └── utils.ts                ← Utilitaires
├── package.json
└── Dockerfile
```

### Règles
- **TypeScript strict** : `strict: true`, JAMAIS `any`
- **Nommage fichiers** : `kebab-case` (memory-star.tsx)
- **Nommage composants** : `PascalCase` (MemoryStar)
- **Imports** : absolus avec `@/` alias
- **"use client"** : UNIQUEMENT sur les composants interactifs (3D, forms, state)
- **Server Components** : par défaut, tout est Server Component sauf besoin explicite
- **Pas de console.log** : utiliser un logger ou supprimer avant commit

---

## SÉCURITÉ — RÈGLES ABSOLUES

### Secrets
- JAMAIS de clé API, token, mot de passe en dur dans le code
- JAMAIS committer `.env`, `.env.*`, `*.key`, `*.pem`
- Toutes les clés via `os.Getenv()` (Go) ou `process.env.` (TS)
- Le fichier `.env.example` montre la structure SANS valeurs

### Inputs
- TOUS les inputs utilisateur validés côté backend (Go: validation custom)
- TOUS les inputs validés côté frontend (Zod schemas)
- Requêtes SQL : TOUJOURS avec paramètres préparés (`?` placeholders), JAMAIS concaténation
- Headers HTTP : validés et sanitisés

### Chiffrement
- Mémoire au repos : chiffrée AES-256-GCM avec `MEMORY_ENCRYPTION_KEY`
- Communication : HTTPS/TLS en production
- Tokens : JWT avec expiration courte (15min access, 7j refresh)
- Mots de passe : hashés avec bcrypt (cost 12+)

### HTTP Security Headers (sur CHAQUE réponse)
```go
Content-Security-Policy: default-src 'self'
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

### CORS
- Origines explicites depuis `CORS_ALLOWED_ORIGINS` env var
- JAMAIS `Access-Control-Allow-Origin: *` en production
- Methods autorisées : GET, POST, DELETE, OPTIONS uniquement

### Rate Limiting
- Toutes les routes publiques : max `RATE_LIMIT_RPM` requêtes/minute/IP
- Routes auth (login/register) : max 10 requêtes/minute/IP
- Routes proxy IA : max 30 requêtes/minute/user

### Docker
- Images basées sur `distroless` ou `alpine` (surface d'attaque minimale)
- `no-new-privileges: true`
- `read_only: true` avec tmpfs pour /tmp
- JAMAIS tourner en root dans le container

---

## BASE DE DONNÉES — SCHÉMA

### Table `users`
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,                    -- UUID v4
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,            -- bcrypt
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Table `conversations`
```sql
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,                    -- UUID v4
    user_id TEXT NOT NULL REFERENCES users(id),
    title TEXT DEFAULT '',
    provider TEXT NOT NULL,                 -- openai | anthropic | ollama
    model TEXT NOT NULL,                    -- gpt-4o | claude-sonnet | llama3
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_conversations_user ON conversations(user_id);
```

### Table `messages`
```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,                    -- UUID v4
    conversation_id TEXT NOT NULL REFERENCES conversations(id),
    role TEXT NOT NULL CHECK(role IN ('system', 'user', 'assistant')),
    content TEXT NOT NULL,                  -- chiffré AES-256
    provider TEXT NOT NULL,
    token_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_messages_conversation ON messages(conversation_id);
```

### Table `memories`
```sql
CREATE TABLE memories (
    id TEXT PRIMARY KEY,                    -- UUID v4
    user_id TEXT NOT NULL REFERENCES users(id),
    summary TEXT NOT NULL,                  -- chiffré AES-256
    source_conversation_ids TEXT NOT NULL,  -- JSON array d'IDs
    theme TEXT DEFAULT '',                  -- thème détecté auto
    importance REAL DEFAULT 0.5,            -- 0.0 à 1.0
    position_x REAL DEFAULT 0.0,           -- position 3D pour le dashboard
    position_y REAL DEFAULT 0.0,
    position_z REAL DEFAULT 0.0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_memories_user ON memories(user_id);
```

---

## API ROUTES (apps/api/)

### Health
- `GET /api/v1/health` → `{"status": "ok", "version": "0.1.0"}`

### Auth
- `POST /api/v1/auth/register` → créer un compte
- `POST /api/v1/auth/login` → obtenir un JWT
- `POST /api/v1/auth/refresh` → rafraîchir le token

### Chat (proxy)
- `POST /api/v1/chat/completions` → proxy vers le provider (streaming SSE)
  - Header `X-MH-Provider: openai | anthropic | ollama`
  - Header `Authorization: Bearer <jwt>`
  - Body : format OpenAI compatible

### Memories
- `GET /api/v1/memories` → liste des mémoires
- `GET /api/v1/memories/:id` → détail
- `DELETE /api/v1/memories/:id` → supprimer
- `GET /api/v1/memories/search?q=...` → recherche

### Conversations
- `GET /api/v1/conversations` → liste
- `GET /api/v1/conversations/:id` → détail avec messages
- `DELETE /api/v1/conversations/:id` → supprimer

---

## ACTIONS INTERDITES

- JAMAIS supprimer des fichiers ou dossiers sans confirmation
- JAMAIS committer ou pusher sans confirmation
- JAMAIS modifier `.env` ou credentials
- JAMAIS exécuter de requêtes SQL destructives (DROP, DELETE sans WHERE)
- JAMAIS installer de dépendances non vérifiées
- JAMAIS de `panic()` en production Go
- JAMAIS de `any` en TypeScript
- JAMAIS de `console.log` de données sensibles
- JAMAIS de wildcard CORS en production

---

## RÈGLES DE TRAVAIL

1. Toujours LIRE un fichier avant de le modifier
2. Tester chaque feature (Go tests + Playwright)
3. Un commit par feature logique
4. Messages de commit en anglais : `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `security:`
5. Chaque fichier Go a son `*_test.go`
6. Chaque composant React est testé ou testable
7. Zéro warning en production
8. Zéro TODO dans le code mergé

---

## FUSION AVEC mh-gdpr-ai.eu — PLAN DÉTAILLÉ

### Contexte
Le repo `mh-gdpr-ai.eu` (chemin local : `C:\Users\Utilisateur\Documents\sovereign-ai-gateway`)
est un projet existant du même auteur (Mahadillah). C'est un gateway IA GDPR-compliant en **Python**.

### Ce que mh-gdpr-ai.eu fait déjà
- **Détection PII temps réel** : emails, IBAN, noms, téléphones, SSN (15+ types)
- **Masking PII** : remplace les données perso par des placeholders sûrs
- **Routing GDPR** : si PII détecté → force le routing vers un provider EU
- **Stack** : Python, Presidio NLP + regex fallback
- **Package** : `pip install mh-gdpr-ai`

### Fichiers clés de mh-gdpr-ai.eu à intégrer
```
sovereign-ai-gateway/
├── sovereign_gateway/
│   ├── gateway.py          ← classe principale SovereignGateway
│   ├── pii/
│   │   ├── detector.py     ← détection PII (Presidio + regex)
│   │   └── masker.py       ← masking des données perso
│   ├── router/
│   │   └── __init__.py     ← SovereignRouter (routing EU/non-EU)
│   └── models/
│       └── __init__.py     ← schemas (Message, RouteResult, SupportedModel)
└── tests/
    ├── test_pii_detector.py
    ├── test_masker.py
    └── test_sovereign_router.py
```

### Comment la fusion se fait (Phase 9 du CLAUDE_GUIDE.md)

**Étape 1** : Copier le code Python de mh-gdpr-ai.eu dans `packages/gdpr-plugin/`
```
packages/gdpr-plugin/
├── sovereign_gateway/     ← copié depuis mh-gdpr-ai.eu
│   ├── gateway.py
│   ├── pii/detector.py
│   ├── pii/masker.py
│   └── router/__init__.py
├── pyproject.toml
├── Dockerfile
└── README.md
```

**Étape 2** : Le plugin GDPR tourne comme un **microservice Python** à côté du backend Go
```
docker-compose.yml:
  api (Go)        ← le protocole principal
  web (Next.js)   ← le frontend
  gdpr (Python)   ← le plugin mh-gdpr-ai.eu, port 8081
```

**Étape 3** : Le backend Go appelle le microservice GDPR
```
Route dans Go :
  1. User envoie un message
  2. Go appelle http://gdpr:8081/detect avec le texte
  3. Si PII détecté → force routing EU (Mistral EU, etc.)
  4. Si pas de PII → routing normal (cheapest)
```

**Étape 4** : Activer/désactiver le plugin via config
```env
GDPR_PLUGIN_ENABLED=true
GDPR_PLUGIN_URL=http://localhost:8081
```

### Quand fusionner
- **PAS au MVP.** Le MVP se concentre sur mémoire + dashboard 3D.
- **Phase 9** (après le buzz) : fusionner mh-gdpr-ai.eu comme plugin.
- Le repo mh-gdpr-ai.eu continue d'exister indépendamment (utilisable seul via `pip install`).
- Ce repo (mh-gdpr-ai.eu-s-plus) l'intègre comme plugin optionnel.

### Liens entre les 2 repos
| | mh-gdpr-ai.eu | mh-gdpr-ai.eu-s-plus (ce repo) |
|---|---|---|
| **Rôle** | Plugin GDPR standalone | Protocole IA complet |
| **Langage** | Python | Go + TypeScript |
| **Relation** | Devient un plugin de ce repo | Intègre mh-gdpr-ai.eu en Phase 9 |
| **GitHub** | github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu | github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus |
