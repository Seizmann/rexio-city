# RexiO City — Product Requirements Document (V1)

**Status:** Active Development  
**Last Updated:** 2026-08-12  
**Author:** SpritEX (Sijan)  
**Version:** v1.1.0

---

## 1. Overview

RexiO City is a public social platform combining Twitter-style short-form posting with Instagram-style profile and media presentation. Built entirely by AI coding agents under human review.

**Live Domain:** `city.rexio.pro`

### Key Value Propositions
- Mobile-first, fast, responsive social platform
- Dark/light mode with OS preference
- Real-time DMs with at-rest encryption
- Feed scoring based on engagement signals
- Admin dashboard for content moderation

---

## 2. Architecture

```
rexio-city/
├── frontend/              # Next.js 16 PWA — city.rexio.pro
├── admin/                 # Vite + Tailwind — oppscity.rexio.pro
├── backend/
│   ├── go/
│   │   ├── cmd/api/       # Main platform backend (port 10800)
│   │   ├── cmd/admin/     # Admin backend (port 10900)
│   │   ├── internal/      # Business logic
│   │   └── migrations/    # SQL migrations
├── docker/                # Docker Compose files
├── Docs/                  # Brand assets (logo, etc.)
├── AGENTS.md              # AI agent instructions
├── DESIGN.md              # Visual design system
├── DECISIONS.md           # Architecture decisions
├── ROADMAP.md             # Project roadmap
├── BRANDING.md            # Brand guidelines
├── TESTING.md             # Testing procedures
├── WORKLOGS.md            # Session logs
├── .env.example           # Environment template
└── Private-Info.md        # Secrets (gitignored)
```

---

## 3. Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 16 (App Router), PWA |
| Admin | Vite + Tailwind CSS |
| Backend | Go (Fiber) |
| Database | Supabase Postgres + PgBouncer |
| Cache | Redis |
| Media | Cloudflare R2 CDN |
| Auth | Custom (argon2id, OAuth2, JWT) |
| DMs | Custom WebSocket server |
| Email | Brevo (transactional) |
| CI/CD | GitHub Actions |
| Hosting | Vercel (frontend) + Cloudflare Tunnel (backend) |

---

## 4. Domains & Deployment

| App | Environment | Domain |
|---|---|---|
| Frontend (Next.js) | Production | city.rexio.pro |
| Frontend (Next.js) | Dev preview | dev-city.rexio.pro |
| Admin (Vite) | Production | oppscity.rexio.pro |
| Admin (Vite) | Local dev | localhost:5189 |
|| Backend (main) | Dev (Cloudflare Tunnel) | devv-connect2city.citydev.rexio.pro |
|| Backend (main) | Production (future VPS) | connect2city.spritexai.dpdns.org |
| Media/CDN | All | cdn-city.rexio.pro |

### Deployment Targets

```bash
# Frontend (Vercel)
Production: https://city.rexio.pro
Preview:    https://dev-city.rexio.pro

# Admin (Vercel)
Production: https://oppscity.rexio.pro

# Backend (Cloudflare Tunnel)
Dev:        https://devv-connect2city.spritexai.dpdns.org
Production: https://connect2city.spritexai.dpdns.org

# Media CDN
All:        https://cdn-city.rexio.pro
```

---

## 5. Feature Requirements

### 5.1 Authentication

**Methods:**
- Email/password with argon2id hashing
- Google OAuth2
- GitHub OAuth2
- Custom JWT + refresh token sessions
- No Supabase Auth — fully custom implementation

**Flow:**
1. User signs up → email verification via Brevo
2. User logs in → JWT issued (short-lived) + refresh token (long-lived)
3. Refresh token rotation on each use
4. Social auth (Google/GitHub) → custom JWT issued

### 5.2 Feed

**Tabs:**
- Following: Posts from followed users only
- For You: Algorithmic feed based on engagement scoring

**Scoring Factors:**
- Recency (decay over time)
- Engagement (likes, comments, reposts)
- Relationship strength (followers, interactions)
- Content type preference

**API:**
- `GET /api/posts/feed?tab=following|foryou&page=N`
- `GET /api/posts/feed/ranking` (scoring weights, DB-driven)

### 5.3 Posts

**Types:**
- Text only (max 500 chars)
- Text + photos (up to 10)
- Text + video (max 5 min, 500MB)
- Text + voice attachment (max 30 sec)
- Link previews (OG metadata fetch)

**Media URLs:**
- Pattern: `cdn-city.rexio.pro/post/{random-id}.{ext}`
- Never sequential/guessable IDs
- R2 CDN with Cloudflare edge caching

**API:**
- `POST /api/posts` — Create post
- `GET /api/posts/:id` — Get post details
- `GET /api/posts/user/:username` — User's posts
- `DELETE /api/posts/:id` — Delete (owner only)

### 5.4 Engagement

**Actions:**
- Like/unlike
- Comment (nested, max 3 levels)
- Repost (with/without comment)
- Bookmark/save

**API:**
- `POST /api/posts/:id/like`
- `GET /api/posts/:id/comments`
- `POST /api/posts/:id/repost`
- `POST /api/posts/:id/bookmark`

### 5.5 Profiles

**Fields:**
- Username (unique, lowercase, 3-15 chars)
- Display name (optional, 1-50 chars)
- Bio (optional, max 160 chars)
- Avatar (uploaded to R2)
- Cover photo (uploaded to R2)
- Join date
- Following/followers count

**Tabs:**
- Posts
- Replies
- Media

### 5.6 Follow System

- Follow/unfollow
- Mutual check
- Follower feed
- Rate limiting: 100 follows/hour

**API:**
- `POST /api/users/:id/follow`
- `GET /api/users/:id/following`
- `GET /api/users/:id/followers`

### 5.7 Direct Messages (DMs)

**Requirements:**
- Real-time via WebSocket
- At-rest encryption (AES-256-GCM)
- NOT end-to-end encrypted (explicitly documented)
- Message history
- Typing indicators
- Read receipts

**WebSocket Events:**
- `message:new` — New DM
- `message:read` — Read receipt
- `typing:start` / `typing:stop` — Typing indicator

**API:**
- `GET /api/dms/conversations`
- `GET /api/dms/:thread/messages`
- `POST /api/dms/:thread/messages`
- `WS /ws/dms` — WebSocket endpoint

### 5.8 Notifications

**Types:**
- New follower
- Like on post
- Comment on post
- Repost
- Mention
- DM reply

**Delivery:**
- In-app notification center
- Email via Brevo (configurable)
- Push notification (PWA)

### 5.9 Search

**Searchable:**
- Users (username, display name)
- Posts (text content)
- Hashtags

**API:**
- `GET /api/search?q=...&type=user|post|hashtag`

---

## 6. API Design

### Base URLs

| Service | Environment | URL |
|---|---|---|
|| Main API | Dev | `https://citydev.rexio.pro` |
|| Main API | Production | `https://connect2city.spritexai.dpdns.org` |
| Media CDN | All | `https://cdn-city.rexio.pro` |

### Response Format
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150
  },
  "error": null
}
```

### Error Format
```json
{
  "success": false,
  "data": null,
  "meta": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Username already taken",
    "details": { "field": "username" }
  }
}
```

---

## 7. Security Requirements

- No direct Supabase from browser — all API calls through Go backend
- CORS: Only allow frontend domains
- Rate limiting on all endpoints
- Input validation and sanitization
- Password hashing: argon2id
- Session: JWT + refresh token rotation
- Media: Signed URLs for upload, random IDs for access
- DM encryption: AES-256-GCM at rest only

---

## 8. Database Schema (High-Level)

```sql
-- Users
users (id, username, display_name, bio, avatar_url, cover_url, created_at, updated_at)

-- Posts
posts (id, user_id, content, created_at, updated_at, deleted_at)

-- Post media
post_media (id, post_id, media_url, media_type, order_index)

-- Engagement
likes (id, user_id, post_id, created_at)
comments (id, user_id, post_id, parent_id, content, created_at)
reposts (id, user_id, post_id, comment, created_at)
bookmarks (id, user_id, post_id, created_at)

-- Follows
follows (follower_id, followee_id, created_at)

-- DMs
dm_conversations (id, created_at)
dm_messages (id, conversation_id, sender_id, encrypted_content, iv, created_at)
dm_participants (conversation_id, user_id)

-- Notifications
notifications (id, user_id, type, actor_id, post_id, read_at, created_at)

-- Settings (DB-driven config)
settings (key, value, updated_at)
```

---

## 9. Admin Features

- User management (view, ban, delete)
- Post moderation (view, delete, hide)
- Media moderation
- System settings (feature flags, rate limits, scoring weights)
- Analytics (user counts, post counts, engagement metrics)

---

## 10. Deployment

### Frontend (Next.js)
- Hosted on Vercel
- Domain: `city.rexio.pro`
- PWA manifest for installability

### Backend (Go)
- Hosted on Hetzner VPS via Cloudflare Tunnel
- Domain: `connect2city.spritexai.dpdns.org`
- Two separate services: main API + admin API

### Admin Panel (Vite)
- Hosted on Vercel
- Domain: `oppscity.rexio.pro`

### Media
- Cloudflare R2 bucket
- CDN: `cdn-city.rexio.pro`

### Database
- Supabase Postgres
- PgBouncer for connection pooling

### Cache
- Redis (Supabase or standalone)

---

## 11. Non-Goals (V1)

- End-to-end encryption for DMs
- Video calling
- Live streaming
- Advanced moderation AI
- Mobile native apps (PWA only)
- Multiple languages (English only)

---

## 12. Success Metrics

- Response time: <200ms for API
- Uptime: 99.9%
- Media upload: <10s for typical photos
- Feed load: <1s for first paint

---

## 13. References

- `AGENTS.md` — Agent workflow rules
- `DESIGN.md` — Visual design system
- `DECISIONS.md` — Architecture decisions
- `BRANDING.md` — Brand guidelines
- `ROADMAP.md` — Project roadmap

---

*This PRD is the binding specification. Any changes must be documented in `WORKLOGS.md` with approval from the project owner.*
