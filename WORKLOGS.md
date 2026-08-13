## [2026-08-13 16:00, GMT+6] — Agent: Hermes (Tanisha) — Model: agnes-2.0-flash
### Picking up:
- User reported navigation delay on repeat visits (Feed → Messages → Alerts → Profile)
- Pages show loading/skeleton state for several seconds even on repeat visits
- Issue: No data caching layer exists — every navigation triggers fresh API calls

### Root cause identified:
- All pages use `useEffect` + `useState` for API calls with no caching
- No SWR or React Query dependency in package.json
- Every tab/page navigation causes fresh network requests
- Skeleton loading state shown until response arrives

### Fix applied:
- Created `src/lib/cache.ts` — in-memory cache with stale-while-revalidate pattern
- Created `src/lib/useCachedFetch.ts` — custom hook for cached data fetching
- Updated `src/app/(main)/page.tsx` (Feed) to use `useCachedFetch`
- Updated `src/app/(main)/[username]/page.tsx` (Profile) to use `useCachedFetch`
- Cache invalidation on post submit and post updates

### Bug fix:
- Fixed cache bug where `getCached` returned `undefined` instead of `null` during fetch
- This caused the feed to show empty state instead of loading skeletons

### Files modified:
- `src/lib/cache.ts` (new)
- `src/lib/useCachedFetch.ts` (new)
- `src/app/(main)/page.tsx` (modified)
- `src/app/(main)/[username]/page.tsx` (modified)

### Notes for next agent:
- Messages and Notifications pages don't exist yet (routes `/messages` and `/notifications` not implemented)
- When those pages are created, apply the same `useCachedFetch` pattern
- Cache TTL is 5 minutes — can be adjusted in `cache.ts` if needed

### Done:
- ✅ Created in-memory cache layer (`cache.ts`) with 5-minute TTL
- ✅ Created `useCachedFetch` hook with stale-while-revalidate pattern
- ✅ Updated Feed page (`/`) to use cached fetch — shows cached data immediately on repeat visits
- ✅ Updated Profile page (`/[username]`) to use cached fetch — same improvement
- ✅ Cache invalidation on post submit and post updates
- ✅ Fixed cache bug causing empty feed
- ✅ TypeScript type-check passes
- ✅ ESLint passes (0 errors, only pre-existing warnings)
- ✅ Next.js build succeeds
- ✅ Commit pushed to `dev` branch (commit `ef0ce7d`)

### Left incomplete:
- Messages page (`/messages`) and Notifications page (`/notifications`) don't exist yet — will need caching when implemented

### Notes for next agent:
- The cache is in-memory only (per-page refresh clears it). This is intentional for V1.
- If backend latency becomes an issue, profile page also does concurrent fetches for user data, follow counts, and posts — consider batching these if needed.

### Secondary bug fix (2026-08-13):
- Fixed issue where cache stored empty data before user authentication
- Added `clearCache()` calls on login, signup, logout, and logoutAll
- This ensures fresh data is fetched after authentication state changes

### Tertiary bug fix (2026-08-13):
- Fixed API response type mismatch for feed endpoint
- Backend returns `data` as `Post[]` directly, not wrapped in `{ data: Post[] }`
- Updated type annotation in feed fetch to match actual backend response

---

## [2026-08-13 17:00, GMT+6] — Agent: Hermes (Tanisha) — Model: agnes-2.0-flash
### Picking up:
- User requested moving dev preview frontend from Vercel to self-hosted on home server
- Reason: Vercel Hobby plan daily deploy limits were exhausted
- Production (city.rexio.pro) stays on Vercel — only dev preview moves

### Tasks completed:
1. Added `start:preview` script to `frontend/package.json` — runs `next start -p 3800`
2. Updated `rexio-city-v1.md` section 4 (Domains & Deployment) — noted dev preview is now self-hosted
3. Updated `rexio-city-v1.md` section 10 (Deployment) — documented both production (Vercel) and dev (self-hosted) frontend
4. Added decision D9 to `DECISIONS.md` — documenting the move, rationale, and reversibility
5. No backend CORS changes needed — `FRONTEND_URL` in backend config covers the dev domain

### Notes for next agent:
- Dev preview frontend now runs on port 3800 on home server
- Cloudflare Tunnel credentials live in ~/.cloudflared/ — NEVER commit them
- Production (city.rexio.pro) and admin (oppscity.rexio.pro) remain on Vercel
- To start dev preview: `npm run build && npm run start:preview`

### Done:
- ✅ Added `start:preview` script to package.json
- ✅ Updated PRD (rexio-city-v1.md) with self-hosted dev preview info
- ✅ Added D9 decision to DECISIONS.md
- ✅ All changes pushed to dev branch (commit `6e23f80`)

### Left incomplete:
- None — this task is complete

### Files modified:
- `frontend/package.json` (added `start:preview` script)
- `rexio-city-v1.md` (updated deployment docs)
- `DECISIONS.md` (added D9 decision)

---

## [2026-08-13 17:30, GMT+6] — Agent: Hermes (Tanisha) — Model: agnes-2.0-flash
### Picking up:
- User requested updating PRD, AGENTS.md, and DECISIONS.md with new rules
- Add rule: NO push without manual user approval

### Tasks completed:
1. Updated `AGENTS.md` Section 2 — added rule: NEVER push without explicit user approval
2. Updated `rexio-city-v1.md` — confirmed deployment docs are current (already updated in previous session)
3. Added D10 to `DECISIONS.md` — Manual Approval Required Before Push
4. Verified tunnel is working at https://dev-city.rexio.pro

### Notes for next agent:
- All pushes to dev branch now require explicit user approval
- Tunnel credentials remain in ~/.cloudflared/ — never commit them
- Frontend dev preview runs on port 3800 via `npm run start:preview`

### Done:
- ✅ Updated AGENTS.md with manual approval rule
- ✅ Added D10 decision to DECISIONS.md
- ✅ Verified tunnel is working

### Left incomplete:
- None — pending user approval to push commits

### Files modified:
- `AGENTS.md` (added manual approval rule)
- `DECISIONS.md` (added D10 decision)

---

## [2026-08-13 18:00, GMT+6] — Agent: Hermes (Tanisha) — Model: agnes-2.0-flash
### Picking up:
- User requested updating all docs to include self-hosted dev preview workflow
- Goal: Ensure no AI agent makes wrong assumptions about deployment in future sessions

### Tasks completed:
1. Updated `AGENTS.md` Section 2.1 — added full Local Dev Preview Workflow
2. Updated `DECISIONS.md` — added D11 documenting self-hosted dev preview
3. Updated `rexio-city-v1.md` — added note about testing on self-hosted preview
4. Documented tunnel setup, testing workflow, and verification steps

### Notes for next agent:
- Dev preview is at https://dev-city.rexio.pro (self-hosted, NOT Vercel)
- Tunnel runs on port 3800, not 3000
- Always test on the public URL, not just localhost
- Tunnel credentials are in ~/.cloudflared/ — never commit them
- Use `systemctl --user status cloudflared` to check tunnel status

### Done:
- ✅ Added Local Dev Preview Workflow to AGENTS.md (Section 2.1)
- ✅ Added D11 decision to DECISIONS.md
- ✅ Updated PRD deployment section with testing note
- ✅ All docs now reflect self-hosted dev preview reality

### Left incomplete:
- None — docs are now complete and accurate

### Files modified:
- `AGENTS.md` (added Section 2.1 with full workflow)
- `DECISIONS.md` (added D11)
- `rexio-city-v1.md` (added testing note in Section 10)

---

## [2026-08-13 18:30, GMT+6] — Agent: Hermes (Tanisha) — Model: agnes-2.0-flash
### Picking up:
- User reported profile showing 0 followers, 0 following, no posts
- Issue: User profile endpoints were behind auth middleware but shouldn't be

### Root cause:
- `GET /api/users/:username` was protected by auth middleware
- `GET /api/users/:id/follow-counts` was protected by auth middleware
- `GET /api/users/:id/is-following` was protected by auth middleware
- These should be PUBLIC so anyone can view profiles without logging in

### Fix applied:
- Moved these three endpoints to public routes (before auth middleware)
- Removed duplicate registrations from protected routes
- Backend now serves public profile data without authentication

### Files modified:
- `backend/go/cmd/api/main.go`

### Status:
- Fix committed but NOT pushed yet (user requested skip)
- Commit: `8fef22b`

### Notes for next agent:
- Push commit `8fef22b` when user gives approval
- Backend build requires Go installed locally or in CI
- Backend runs as root via systemd/service — restart with: `sudo systemctl restart rexio-backend` or manually kill + restart

---

## [2026-08-13 18:45, GMT+6] — Agent: Hermes (Tanisha) — Model: agnes-2.0-flash
### Picking up:
- User reported profile showing 0 followers, 0 following for all users
- Root cause: Profile endpoints were behind auth middleware but should be public

### Root cause identified:
- `GET /api/users/:username` - was protected, should be PUBLIC
- `GET /api/users/:id/follow-counts` - was protected, should be PUBLIC
- `GET /api/users/:id/is-following` - was protected, should be PUBLIC

### Fix applied:
- Moved these 3 endpoints to public routes (before auth middleware)
- Removed duplicate registrations from protected routes
- Commit: `8fef22b`

### Current status:
- ✅ Backend code fixed (commit `8fef22b` pushed to dev)
- ✅ Frontend running on port 3800 (dev-city.rexio.pro)
- ❌ Backend NOT restarted yet — still running old binary (PID 1975724 as root)
- ⚠️ Backend runs as root via systemd, needs sudo to restart

### To restart backend:
```bash
# Option 1: Using systemd (if service exists)
sudo systemctl restart rexio-backend

# Option 2: Manual restart
sudo pkill -f "rexio-city/backend/go/api"
cd /home/sijan/SijansP/rexio-city/backend/go
sudo -u sijan DATABASE_URL="$(grep DATABASE_URL .env | cut -d= -f2-)" nohup ./api > /tmp/rexio-backend.log 2>&1 &
```

### Files modified:
- `backend/go/cmd/api/main.go`
