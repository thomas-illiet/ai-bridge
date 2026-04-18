# AI Bridge

> One bridge. Every model. Infinite possibilities.

AI Bridge is a self-hosted proxy that sits between your users and LLM providers (OpenAI, Ollama, and more). It enforces authentication and role-based access, records every request for auditing and analytics, and exposes an OpenAI-compatible API so any existing client works without modification.

---

## Features

- **Unified AI proxy** — transparent pass-through to OpenAI and Ollama behind a single authenticated endpoint
- **Keycloak SSO** — all access is gated through OpenID Connect; JWT and personal access tokens (PAT) both supported
- **Role-based access control** — three roles (`none`, `user`, `admin`) with optional expiry dates
- **Access request workflow** — users request access; admins approve or reject with optional role expiry
- **IP whitelisting** — optionally restrict proxy access to known CIDRs
- **Usage tracking** — every interception is recorded: tokens consumed, prompts, tool calls, model thoughts
- **Admin dashboard** — manage users, tokens, history, whitelist, and access requests from a single panel
- **User dashboard** — personal stats, token history, request charts, and model breakdown
- **Floating AI chat widget** — in-browser chat with streaming responses, provider/model selection, and conversation history
- **Prometheus metrics + OpenTelemetry tracing** — production-grade observability out of the box
- **Email notifications** — SMTP alerts for access approvals and rejections

---

## Architecture

```
┌──────────────┐     JWT / PAT      ┌─────────────────────────────────┐
│   Browser    │ ─────────────────► │          Go API (Gin)           │
│  Vue 3 SPA   │                    │  /api/v1/*   — REST endpoints    │
└──────────────┘                    │  /openai/*   — OpenAI proxy      │
                                    │  /ollama/*   — Ollama proxy       │
                                    └──────────┬──────────────────────┘
                                               │
                    ┌──────────────────────────┼───────────────────────┐
                    ▼                          ▼                       ▼
             PostgreSQL                    Keycloak               OpenAI / Ollama
           (GORM models)               (OIDC provider)            (upstream LLMs)
```

---

## Quick Start (Docker)

**Prerequisites:** Docker and Docker Compose.

```bash
git clone https://github.com/your-org/ai-bridge.git
cd ai-bridge

# Optional: set your OpenAI key
export OPENAI_API_KEY=sk-...

docker compose up -d
```

Services start in dependency order (postgres → keycloak → backend → frontend). Wait ~30 s for Keycloak to finish importing the realm, then open:

| Service | URL |
|---|---|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:8585 |
| Keycloak admin | http://localhost:8180 |
| Maildev (email UI) | http://localhost:1080 |
| Ollama | http://localhost:11434 |

The default Keycloak admin credentials are `admin` / `admin`. The realm `ai-bridge` is imported automatically.

---

## Local Development

### Backend

```bash
cd backend
cp .env.example .env   # edit as needed
go run .
```

Requires Go ≥ 1.22, a running PostgreSQL instance, and a reachable Keycloak.

### Frontend

```bash
cd frontend
cp .env.example .env   # edit VITE_* values if needed
npm install
npm run dev
```

The Vite dev server starts on `http://localhost:5173` and proxies `/api` to the backend.

---

## Configuration

### Backend environment variables

| Variable | Default | Required | Description |
|---|---|---|---|
| `DATABASE_DSN` | — | ✅ | PostgreSQL connection string |
| `TOKEN_SECRET` | — | ✅ | Secret used to sign PATs (≥ 32 chars) |
| `KEYCLOAK_BASE_URL` | `http://localhost:8180` | ✅ | Keycloak server URL |
| `KEYCLOAK_REALM` | `ai-bridge` | ✅ | Keycloak realm name |
| `KEYCLOAK_CLIENT_ID` | `ai-bridge-frontend` | | OIDC client ID |
| `KEYCLOAK_ISSUER_URL` | same as base URL | | Override issuer URL (useful behind reverse proxies) |
| `SERVER_PORT` | `8585` | | HTTP listen port |
| `ALLOWED_ORIGINS` | `http://localhost:5173` | | Comma-separated CORS origins |
| `OPENAI_API_KEY` | — | | Enables OpenAI proxy when set |
| `OLLAMA_BASE_URL` | — | | Enables Ollama proxy when set (e.g. `http://localhost:11434`) |
| `OLLAMA_NUM_CTX` | `4096` | | Ollama context window size |
| `TRUSTED_PROXIES` | RFC-1918 ranges | | Comma-separated CIDRs of trusted reverse proxies |
| `SMTP_HOST` | — | | SMTP server host (email notifications) |
| `SMTP_PORT` | `587` | | SMTP port |
| `SMTP_USER` | — | | SMTP username |
| `SMTP_PASSWORD` | — | | SMTP password |
| `SMTP_FROM` | — | | Sender email address |
| `SMTP_TO` | — | | Comma-separated admin email addresses |
| `APP_URL` | `http://localhost:5173` | | Public base URL used in email links |
| `ROLE_EXPIRY_INTERVAL_SEC` | `60` | | How often (seconds) to sweep expired roles |

### Frontend environment variables

| Variable | Default | Description |
|---|---|---|
| `VITE_KEYCLOAK_URL` | `http://localhost:8180` | Keycloak server URL |
| `VITE_KEYCLOAK_REALM` | `ai-bridge` | Keycloak realm |
| `VITE_KEYCLOAK_CLIENT_ID` | `ai-bridge-frontend` | OIDC client ID |
| `VITE_API_BASE_URL` | `http://localhost:8585` | Backend base URL (used for AI proxy calls) |

---

## API Overview

All API routes (except `/health`, `/api/status`, `/metrics`) require a valid Bearer token — either a Keycloak JWT or a personal access token.

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/api/status` | Available providers and their status |
| `GET` | `/metrics` | Prometheus metrics |

### User (role: `user` or `admin`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/me` | Current user profile |
| `GET` | `/api/v1/dashboard` | Personal usage statistics |
| `GET` | `/api/v1/models` | Available models per provider |
| `POST` | `/api/v1/access-requests` | Submit an access request |
| `GET` | `/api/v1/access-requests/me` | Own access request status |
| `POST` | `/api/v1/tokens` | Create a personal access token |
| `GET` | `/api/v1/tokens` | List own tokens |
| `DELETE` | `/api/v1/tokens/:id` | Revoke a token |
| `GET` | `/api/v1/history` | Own request history |
| `GET` | `/api/v1/history/:id` | Request detail |
| `ANY` | `/openai/*` | OpenAI-compatible proxy |
| `ANY` | `/ollama/*` | Ollama proxy |

### Admin (role: `admin`)

| Method | Path | Description |
|---|---|---|
| `GET/POST` | `/api/v1/admin/whitelist` | Manage IP whitelist |
| `DELETE/PATCH` | `/api/v1/admin/whitelist/:id` | Remove / toggle entry |
| `GET/PATCH/DELETE` | `/api/v1/admin/users` | Manage users and roles |
| `GET` | `/api/v1/admin/tokens` | All users' tokens |
| `DELETE` | `/api/v1/admin/tokens/:id` | Revoke any token |
| `GET` | `/api/v1/admin/history` | All users' request history |
| `GET/POST` | `/api/v1/admin/access-requests` | List / approve / reject requests |

---

## Using the Proxy

The proxy endpoints are OpenAI-compatible. Point any OpenAI SDK or tool at AI Bridge instead:

```bash
# cURL — Ollama via AI Bridge
curl http://localhost:8585/ollama/v1/chat/completions \
  -H "Authorization: Bearer <your-pat>" \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3.2","messages":[{"role":"user","content":"Hello!"}],"stream":false}'

# Python (openai SDK) — OpenAI via AI Bridge
from openai import OpenAI
client = OpenAI(api_key="<your-pat>", base_url="http://localhost:8585/openai/v1")
response = client.chat.completions.create(model="gpt-4o", messages=[...])
```

---

## Project Structure

```
ai-bridge/
├── backend/
│   ├── aibridge/          # Ollama provider + GORM usage recorder
│   ├── config/            # Environment-based configuration
│   ├── database/          # PostgreSQL connection (GORM)
│   ├── handlers/          # HTTP handlers
│   ├── middleware/         # JWT auth, RBAC, IP whitelist
│   ├── models/            # GORM models
│   ├── services/          # Business logic (tokens, users, email)
│   ├── main.go
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/    # Shared UI components
│   │   ├── composables/   # Vue composition utilities
│   │   ├── router/        # Vue Router routes
│   │   ├── services/      # Axios API client
│   │   ├── stores/        # Pinia state stores
│   │   └── views/         # Page components (admin/, chat/, dashboard/, …)
│   ├── .env.example
│   └── Dockerfile
├── keycloak/
│   └── ai-bridge-realm.json   # Pre-configured realm (auto-imported)
└── docker-compose.yml
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go · Gin · GORM · PostgreSQL |
| Authentication | Keycloak (OIDC) · golang-jwt |
| AI proxy | github.com/coder/aibridge |
| Frontend | Vue 3 · TypeScript · Pinia · Vite |
| Observability | Prometheus · OpenTelemetry |
| Infrastructure | Docker Compose · Distroless container images |

---

## License

MIT
