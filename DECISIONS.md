# DECISIONS.md — Architecture Decisions

Record of binding architecture decisions for the RexiO City project. Do not silently reverse these — if a new decision conflicts, document the change and get explicit approval.

---

## D1: No Direct Supabase from Browser

**Decision:** All data access goes through the Go backend. Neither `frontend/` nor `admin/` calls Supabase directly.

**Rationale:** Centralizes auth, rate limiting, and business logic. Prevents leaking DB schema or credentials. Enables consistent error handling and audit logging.

**Status:** Active  
**Date:** 2026-08-12

---

## D2: Custom Auth (No Supabase Auth)

**Decision:** Full custom auth with argon2id password hashing, custom OAuth2 for Google/GitHub, and JWT + refresh token sessions.

**Rationale:** Gives full control over session management, token rotation, and social login flow. Avoids Supabase Auth limitations.

**Status:** Active  
**Date:** 2026-08-12

---

## D3: Separate Admin Backend

**Decision:** `cmd/admin` is a separate Go binary with its own Supabase connection (admin project). Main platform and admin share Postgres schemas but use different credentials.

**Rationale:** Blast radius isolation. A compromise of the admin backend doesn't expose user data directly. Allows different rate limits and auth requirements.

**Status:** Active  
**Date:** 2026-08-12

---

## D4: DMs Use Custom WebSocket Server

**Decision:** Real-time messaging uses a custom Go WebSocket server, not Supabase Realtime.

**Rationale:** Supabase Realtime doesn't support the encryption model needed. Custom server allows fine-grained control over message delivery, typing indicators, and presence.

**Status:** Active  
**Date:** 2026-08-12

---

## D5: Cloudflare R2 for Media Storage

**Decision:** All media (photos, videos, avatars) stored in Cloudflare R2, served via CDN (`cdn-city.rexio.pro`).

**Rationale:** Cost-effective object storage with global CDN. Integrates well with Cloudflare Tunnel for backend. Signed URLs enable secure direct uploads.

**Status:** Active  
**Date:** 2026-08-12

---

## D6: Presigned URL Upload Flow

**Decision:** Large media uploads use presigned URL flow: frontend requests URL from backend, then uploads directly to R2, then confirms with backend.

**Rationale:** Bypasses Vercel's 4.5MB function payload limit. Allows mobile camera photos (5-15MB) to upload successfully. Client-side compression reduces bandwidth.

**Status:** Active  
**Date:** 2026-08-13

---

## D7: DB-driven Configuration

**Decision:** Business logic configuration (feature flags, rate limits, scoring weights) stored in database `settings` table, not hardcoded `.env` values.

**Rationale:** Allows runtime changes without redeployment. Enables A/B testing and gradual rollouts.

**Status:** Active  
**Date:** 2026-08-12

---

## D8: Go Backend as Single API Gateway

**Decision:** All frontend requests go through Go backend. No direct Supabase from browser.

**Rationale:** Security, consistency, and centralized business logic. The backend is the source of truth.

**Status:** Active  
**Date:** 2026-08-12

---

*Record all future decisions here with date, status, and rationale.*
