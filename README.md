# RexiO City

A public social platform combining Twitter-style short-form posting with Instagram-style profile and media presentation. Built entirely by AI coding agents under human review.

**Live:** [city.rexio.pro](https://city.rexio.pro)

---

## Overview

RexiO City is a general-purpose social platform featuring:

- **Feed** — Following + For You tabs with engagement-based scoring
- **Posts** — Text, photos, videos, voice attachments with link previews
- **Profiles** — Username, bio, avatar, cover photo, media tabs
- **Engagement** — Likes, comments, reposts, bookmarks
- **DMs** — Real-time messaging via WebSocket with at-rest encryption
- **Auth** — Custom OAuth2 (Google, GitHub) + password auth with argon2id
- **Admin** — Separate dashboard for moderation and user management

## Architecture

```
rexio-city/
├── frontend/              # Next.js PWA — city.rexio.pro
├── admin/                 # Vite + Tailwind — separate admin app
├── backend/
│   ├── go/
│   │   ├── cmd/api/       # Main platform backend
│   │   └── cmd/admin/     # Admin backend (separate)
│   └── migrations/        # SQL migrations
├── docker/
├── Docs/                  # Brand assets
├── AGENTS.md              # AI agent instructions
├── DESIGN.md              # Visual design system
├── rexio-city-v1.md       # Product requirements
└── WORKLOGS.md            # Session logs
```

### Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 16 (App Router), PWA |
| Admin | Vite + Tailwind |
| Backend | Go (Gin/Fiber) |
| Database | Supabase Postgres + PgBouncer |
| Cache | Redis |
| Media | Cloudflare R2 CDN |
| Auth | Custom (argon2id, OAuth2, JWT) |
| DMs | Custom WebSocket server |
| Email | Brevo (transactional) |
| CI/CD | GitHub Actions |
| Hosting | Vercel (frontend) + Cloudflare Tunnel (backend) |

## Key Design Decisions

- **No direct Supabase from browser** — All API calls go through Go backend
- **DB-driven configuration** — Feature flags, rate limits, scoring weights in database
- **Separate admin backend** — Blast-radius isolation from main platform
- **At-rest DM encryption** — AES-256-GCM (not end-to-end)
- **Single-agent-at-a-time** — Sequential AI agent development, manual PR merges

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker + Docker Compose
- Supabase project (main + admin)
- Cloudflare R2 bucket
- Redis instance

### Local Development

```bash
# Clone and setup
git clone git@github.com:Seizmann/rexio-city.git
cd rexio-city

# Copy env files
cp .env.example .env
# Edit .env with your secrets

# Start services
docker compose -f docker/docker-compose.dev.yml up -d

# Frontend
cd frontend && npm install && npm run dev

# Backend
cd ../backend/go && go mod download && go run cmd/api/main.go
```

### Branching

- `main` — Production (protected, manual merges only)
- `dev` — Active development

All AI agents push to `dev`. Sijan reviews and merges to `main`.

## AI-Agent Development

This project is built entirely by AI coding agents (Claude Code, Cursor, OpenCode/Hermes, Antigravity). See `AGENTS.md` for the full agent workflow.

- Every session logs to `WORKLOGS.md`
- Critical logic requires unit tests
- gitleaks CI blocks secret commits
- No fabricated human co-authorship on commits

## License

This project is licensed under the GNU Affero General Public License v3.0 — see [LICENSE](LICENSE).
