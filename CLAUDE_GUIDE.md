# mh-gdpr-ai.eu S+ — Guide de Construction

> Du zéro au MVP au produit complet.
> Chaque phase a un CHECKPOINT. Ne passe pas à la suivante tant que le checkpoint n'est pas validé.

---

## VUE D'ENSEMBLE

```
PHASE 0 — Setup repo + configs + Docker           → tout tourne à vide
PHASE 1 — Proxy IA (OpenAI, Anthropic, Ollama)    → on peut appeler les 3 IA via le proxy
PHASE 2 — Mémoire (SQLite, sauvegarde auto)       → chaque conversation est sauvée
PHASE 3 — Injection contexte (le WOW technique)   → switch de provider sans perte
PHASE 4 — Frontend chat                           → site web fonctionnel
PHASE 5 — Dashboard spatial 3D (le WOW visuel)    → étoiles, flux, espace
PHASE 6 — Polish + Tests + Deploy                 → MVP prêt pour le buzz
─────────────── MVP TERMINÉ ───────────────
PHASE 7+  — Version complète (routeur, multi-IA, GDPR, SDKs...)
```

---

## PHASE 0 — SETUP

### 0.1 : Structure du monorepo
**Déjà fait** : dossiers créés, git init, .gitignore, configs.

### 0.2 : Backend Go — Init
**Fichiers à créer :**
- `apps/api/go.mod` — module Go
- `apps/api/cmd/server/main.go` — entrypoint serveur Gin
- `apps/api/internal/config/config.go` — charge les variables d'env
- `apps/api/internal/middleware/security.go` — CORS, headers, rate limit
- `apps/api/internal/middleware/logger.go` — logging structuré
- `apps/api/Dockerfile` — image Docker multi-stage

**Ce que ça fait :**
```
go run ./cmd/server → serveur Gin démarre sur :8080
GET /api/v1/health → {"status":"ok","version":"0.1.0"}
```

**Sécurité :**
- Security headers sur toutes les réponses
- CORS restrictif
- Rate limiting actif
- Logging sans données sensibles

### 0.3 : Frontend Next.js — Init
**Commande :**
```bash
cd apps/web
npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir=false
```

**Packages à installer :**
```bash
npm install @react-three/fiber @react-three/drei @react-three/postprocessing three
npm install zustand zod framer-motion
npx shadcn@latest init
```

**Fichiers à créer/modifier :**
- `apps/web/app/layout.tsx` — layout avec meta, fonts
- `apps/web/app/page.tsx` — landing page basique
- `apps/web/lib/api.ts` — client API vers backend
- `apps/web/lib/stores/chat-store.ts` — Zustand store pour le chat
- `apps/web/lib/stores/memory-store.ts` — Zustand store pour les mémoires
- `apps/web/lib/schemas.ts` — schémas Zod
- `apps/web/Dockerfile` — image Docker multi-stage

### 0.4 : Docker Compose
**Déjà fait** : docker-compose.yml créé.

**Test :**
```bash
docker-compose up --build
# → API sur localhost:8080
# → Web sur localhost:3000
```

✅ **CHECKPOINT 0** : `docker-compose up` lance les 2 apps. Health check OK.

---

## PHASE 1 — LE PROXY IA

### 1.1 : Format unifié
**Fichier** : `apps/api/internal/proxy/types.go`

Définir les structs internes :
```
ChatRequest  { Messages, Provider, Model, Stream, UserID }
ChatResponse { Content, Provider, Model, TokensUsed }
```

Tous les providers sont convertis vers/depuis ce format.

### 1.2 : Adaptateur OpenAI
**Fichier** : `apps/api/internal/proxy/openai.go`

- Reçoit un `ChatRequest`
- Convertit en format OpenAI API
- Appelle `api.openai.com/v1/chat/completions`
- Supporte le streaming SSE
- Convertit la réponse en `ChatResponse`
- SÉCURITÉ : la clé API vient de `os.Getenv("OPENAI_API_KEY")`, jamais en dur

### 1.3 : Adaptateur Anthropic
**Fichier** : `apps/api/internal/proxy/anthropic.go`

- Même chose mais vers `api.anthropic.com/v1/messages`
- Convertit le format (Anthropic utilise un format différent d'OpenAI)
- Gère le header `x-api-key` et `anthropic-version`
- Streaming via SSE

### 1.4 : Adaptateur Ollama
**Fichier** : `apps/api/internal/proxy/ollama.go`

- Même chose mais vers `localhost:11434/api/chat`
- Pas de clé API (local)
- Convertit le format Ollama en format unifié

### 1.5 : Proxy handler
**Fichier** : `apps/api/internal/proxy/handler.go`

- Route `POST /api/v1/chat/completions`
- Lit le header `X-MH-Provider`
- Dispatch vers le bon adaptateur
- Retourne la réponse en streaming

### 1.6 : Tests proxy
**Fichiers** : `*_test.go` pour chaque adaptateur
- Test de conversion format
- Test de streaming
- Test d'erreur (provider down, clé invalide)

✅ **CHECKPOINT 1** : `curl -X POST localhost:8080/api/v1/chat/completions` avec header provider → réponse IA en streaming.

---

## PHASE 2 — LA MÉMOIRE

### 2.1 : Init SQLite
**Fichier** : `apps/api/internal/memory/sqlite.go`

- Ouvre/crée le fichier `data/mh-gdpr.db`
- Crée les tables (users, conversations, messages, memories) si elles n'existent pas
- SÉCURITÉ : paramètres WAL mode pour la performance, busy timeout

### 2.2 : Modèles
**Fichier** : `apps/api/internal/memory/models.go`

- Structs Go : `User`, `Conversation`, `Message`, `Memory`
- Fonctions de validation pour chaque struct
- Fonctions de chiffrement/déchiffrement AES-256 pour le contenu

### 2.3 : Store interface
**Fichier** : `apps/api/internal/memory/store.go`

```go
type Store interface {
    SaveConversation(ctx context.Context, conv *Conversation) error
    SaveMessage(ctx context.Context, msg *Message) error
    GetConversation(ctx context.Context, id string) (*Conversation, error)
    GetMessages(ctx context.Context, conversationID string) ([]*Message, error)
    ListConversations(ctx context.Context, userID string) ([]*Conversation, error)
    DeleteConversation(ctx context.Context, id string) error
    SaveMemory(ctx context.Context, mem *Memory) error
    GetMemories(ctx context.Context, userID string) ([]*Memory, error)
    SearchMemories(ctx context.Context, userID, query string) ([]*Memory, error)
    DeleteMemory(ctx context.Context, id string) error
}
```

### 2.4 : Auto-save dans le proxy
Modifier `proxy/handler.go` :
- Après chaque requête/réponse → sauvegarder dans la DB automatiquement
- L'utilisateur ne fait rien, tout est transparent

### 2.5 : Résumé automatique
**Fichier** : `apps/api/internal/memory/summarizer.go`

- Toutes les 10 messages dans une conversation → demander un résumé à l'IA
- Stocker le résumé dans la table `memories`
- Le résumé est ce qui sera injecté lors d'un switch
- SÉCURITÉ : le résumé est chiffré au repos

### 2.6 : API endpoints mémoire
**Fichier** : `apps/api/internal/memory/handler.go`

- `GET /api/v1/memories` → liste
- `GET /api/v1/memories/:id` → détail
- `DELETE /api/v1/memories/:id` → supprimer
- `GET /api/v1/memories/search?q=` → recherche
- `GET /api/v1/conversations` → liste
- `GET /api/v1/conversations/:id` → détail + messages
- `DELETE /api/v1/conversations/:id` → supprimer
- SÉCURITÉ : chaque endpoint vérifie que l'user est propriétaire (JWT → user_id)

### 2.7 : Tests mémoire
- Test CRUD conversations
- Test CRUD memories
- Test chiffrement/déchiffrement
- Test résumé auto
- Test search

✅ **CHECKPOINT 2** : les conversations sont sauvées en SQLite, chiffrées, et requêtables via API.

---

## PHASE 3 — INJECTION DE CONTEXTE

### 3.1 : Injecteur
**Fichier** : `apps/api/internal/injector/injector.go`

- Quand l'user switch de provider :
  1. Récupérer les dernières mémoires de l'user
  2. Construire un message `system` avec les résumés
  3. L'ajouter au début de la requête vers le nouveau provider
- Le nouveau provider "sait" tout ce que l'ancien savait

### 3.2 : Détection de switch
Modifier `proxy/handler.go` :
- Tracker le dernier provider utilisé par user
- Si provider actuel ≠ dernier provider → déclencher l'injection

### 3.3 : Tests injection
- Test : envoyer messages à OpenAI, switch Anthropic, vérifier que le contexte est injecté
- Test : le résumé injecté ne dépasse pas X tokens (optimisation coût)

✅ **CHECKPOINT 3** : switch GPT → Claude → Llama, chaque IA connaît le contexte précédent. **C'est le WOW technique.**

---

## PHASE 4 — FRONTEND CHAT

### 4.1 : Landing page
**Fichier** : `apps/web/app/page.tsx`

- Hero section : titre + tagline + vidéo/animation de démo
- Boutons : "Try it" → /chat, "See your memory" → /space, "GitHub" → repo
- Design sombre, spatial, minimaliste

### 4.2 : Auth pages
**Fichiers** : `apps/web/app/(auth)/login/page.tsx`, `register/page.tsx`

- Formulaires avec validation Zod
- Stockage JWT dans httpOnly cookie (PAS localStorage)
- SÉCURITÉ : CSRF protection

### 4.3 : Chat interface
**Fichier** : `apps/web/app/chat/page.tsx`

**Composants :**
- `components/chat/chat-input.tsx` — input avec envoi
- `components/chat/chat-message.tsx` — affichage d'un message (markdown)
- `components/chat/provider-switch.tsx` — boutons GPT / Claude / Llama
- `components/chat/conversation-sidebar.tsx` — liste des conversations

**Fonctionnalités :**
- Envoi de message → appel API → affichage streaming
- Switch de provider en 1 clic (change le header)
- Indicateur visuel du provider actif (couleur : bleu GPT, orange Claude, vert Llama)
- Sidebar avec historique des conversations
- Markdown rendering (code blocks, listes, etc.)

### 4.4 : Zustand stores
- `lib/stores/chat-store.ts` — messages, provider actif, conversation active
- `lib/stores/memory-store.ts` — liste des mémoires, chargement
- `lib/stores/auth-store.ts` — user, token, login/logout

### 4.5 : Client API
**Fichier** : `apps/web/lib/api.ts`

- Classe `MHAPI` avec méthodes typées
- Gestion des erreurs centralisée
- Gestion du streaming SSE
- Auto-refresh du JWT
- SÉCURITÉ : jamais de token dans les URL, toujours dans les headers

✅ **CHECKPOINT 4** : site web fonctionnel, on peut chatter avec GPT/Claude/Llama et switcher.

---

## PHASE 5 — DASHBOARD SPATIAL 3D

### 5.1 : Scène de base
**Fichier** : `apps/web/components/space/memory-universe.tsx`

- Canvas React Three Fiber
- Fond noir (#000000)
- Étoiles de fond (drei Stars)
- Caméra orbite (OrbitControls)
- Éclairage ambiant doux
- Post-processing : Bloom (glow), Vignette

### 5.2 : Étoiles mémoire
**Fichier** : `apps/web/components/space/memory-star.tsx`

- Chaque mémoire = une sphère (SphereGeometry)
- Matériau : MeshStandardMaterial avec emissive (brille)
- Couleur par provider :
  - Bleu #4A90D9 → OpenAI
  - Orange #E87B35 → Anthropic/Claude
  - Vert #4ADE80 → Ollama/Llama
- Taille proportionnelle à l'importance (0.5 → 2.0)
- Opacité proportionnelle à la récence (récent = brillant, ancien = estompé)
- Animation : flottement sinusoïdal lent (sin/cos sur le temps)
- Hover : scale up + tooltip avec le résumé
- Click : panneau latéral avec la conversation complète

### 5.3 : Connexions entre étoiles
**Fichier** : `apps/web/components/space/star-connections.tsx`

- Lines entre mémoires du même thème
- Matériau : LineBasicMaterial avec opacity 0.2
- Plus le lien est fort → plus la ligne est opaque

### 5.4 : Flux de contexte
**Fichier** : `apps/web/components/space/context-flow.tsx`

- Quand un switch de provider est détecté :
  - Particules lumineuses voyagent d'une étoile à l'autre
  - Couleur = gradient du provider source → provider cible
  - Durée : 2 secondes d'animation
- Utilise drei `Trail` ou custom particle system

### 5.5 : Contrôles caméra
**Fichier** : `apps/web/components/space/camera-controls.tsx`

- OrbitControls : rotation, zoom, pan
- Auto-rotate lent quand idle
- Zoom sur un cluster au click
- Smooth transitions (lerp)
- Reset view button

### 5.6 : Page espace
**Fichier** : `apps/web/app/space/page.tsx`

- Full screen canvas
- Overlay UI : bouton retour, search, filtres
- Panneau latéral pour le détail d'une mémoire
- Chargement des mémoires depuis l'API

### 5.7 : Polish visuel
- Post-processing : Bloom, Vignette, ChromaticAberration subtile
- Particules de fond (poussière d'étoiles, lent)
- Transitions de caméra fluides
- Loading state (étoiles apparaissent une par une)
- Mode plein écran

✅ **CHECKPOINT 5** : le dashboard 3D est magnifique, interactif, et connecté aux vraies données. **C'EST LE WOW VISUEL.**

---

## PHASE 6 — MVP LAUNCH

### 6.1 : Tests
- Go : `go test ./...` — tous les packages
- Playwright : test flow complet (register → chat → switch → space)
- k6 : test de charge (100 users simultanés)

### 6.2 : README.md
- Logo + nom + tagline
- GIF animé du dashboard 3D (30 secondes)
- Quick start : 3 commandes
- Screenshots
- Architecture diagram
- Badges GitHub
- Contributing guide
- License

### 6.3 : Déploiement
- Frontend → Vercel
- Backend → Fly.io (ou Railway)
- CI/CD → GitHub Actions (test + build + deploy)
- HTTPS partout
- Domain custom

### 6.4 : Démo vidéo
- 30 secondes, zéro parole, juste visuel + texte
- Poster sur : YouTube, X, Reddit (r/programming, r/artificial), LinkedIn, HackerNews

✅ **MVP TERMINÉ — PRÊT POUR LE BUZZ**

---

## PHASE 7+ — VERSION COMPLÈTE (après le buzz)

### Phase 7 : Routeur intelligent
- Analyse la complexité de la question
- Route auto vers cheapest / fastest / best
- Dashboard des coûts

### Phase 8 : Multi-IA collaboration
- Plusieurs modèles sur une même tâche
- GPT génère → Claude vérifie → résultat fusionné

### Phase 9 : Plugin GDPR (fusion mh-gdpr-ai.eu)
- Importer le code PII detection depuis mh-gdpr-ai.eu
- Activer/désactiver via settings
- Si PII détecté → route vers EU provider

### Phase 10 : Identité agent
- Chaque agent a une identité signée (clé publique/privée)
- Permissions granulaires
- Audit trail complet

### Phase 11 : Vérification des outputs
- Signature cryptographique de chaque réponse
- Traçabilité : quel modèle, quand, quel contexte

### Phase 12 : Multi-device
- Sync mémoire chiffrée entre appareils
- E2E encryption

### Phase 13 : SDKs
- `pip install mh-gdpr-ai` (Python)
- `npm install mh-gdpr-ai` (TypeScript)
- `go get mh-gdpr-ai` (Go)

### Phase 14 : Dashboard avancé
- Filtres par date, provider, thème
- Timeline view
- Search full-text
- Export des données (JSON, CSV)
- Partage de vues
