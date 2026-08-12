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

## D5: DM Encryption is At-Rest Only

**Decision:** DMs are encrypted at rest using AES-256-GCM. NOT end-to-end encrypted.

**Rationale:** E2E encryption requires key exchange protocol that's out of scope for V1. At-rest encryption protects against DB breaches. Product explicitly states "not E2EE" to avoid misrepresentation.

**Status:** Active  
**Date:** 2026-08-12

---

## D6: DB-Driven Configuration

**Decision:** Feature flags, rate limits, and feed scoring weights are stored in a `settings` table, not hardcoded in `.env`.

**Rationale:** Allows changing behavior without redeployment. Enables A/B testing. `.env` is reserved for static secrets.

**Status:** Active  
**Date:** 2026-08-12

---

## D7: Media URLs Use Random IDs

**Decision:** Media URLs follow pattern `cdn-city.rexio.pro/post/{random-id}.{ext}`. No sequential or guessable IDs.

**Rationale:** Prevents enumeration attacks. Random IDs are generated with sufficient entropy (e.g., UUID v4 or 32-char hex).

**Status:** Active  
**Date:** 2026-08-12

---

## D8: Single Agent at a Time

**Decision:** Only one AI coding agent works on this repository in a single session. No concurrent edits.

**Rationale:** Prevents merge conflicts, ensures consistent code quality, allows each agent to read full context before working.

**Status:** Active  
**Date:** 2026-08-12

---

## D9: Branching Model

**Decision:** All work goes to `dev`. `main` is protected and only updated via manual PR review by Sijan.

**Rationale:** Clean history, easy rollback, Sijan has final say on production code.

**Status:** Active  
**Date:** 2026-08-12

---

## D10: Design Token System

**Decision:** All colors, spacing, and typography come from CSS custom properties defined in `DESIGN.md`. No hardcoded values in components.

**Rationale:** Consistency across frontend and admin apps. Easy theme changes. Single source of truth.

**Status:** Active  
**Date:** 2026-08-12

---

*When adding new decisions, append a new section with D-number, decision, rationale, status, and date.*
