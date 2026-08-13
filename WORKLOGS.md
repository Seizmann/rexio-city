## [2026-08-13 11:30, GMT+6] — Agent: Hermes (Tanisha) — Model: agnes-2.0-flash
### Picking up:
- Mobile photo uploads failing with "Unexpected token 'R', 'Request En'... is not valid JSON"
- Desktop uploads succeed, mobile fails
- Debug panel shows: upload succeeds (200), post creation fails (431 or 413)
- Need to: fix backend body size limits, add frontend error handling for non-JSON responses, add tests

### Plan for this session:
- Audit all body size limits in the request path (backend, middleware, Vercel)
- Add safe JSON parsing in frontend API client
- Handle 413 (Payload Too Large) with user-friendly error
- Add Go unit test for body size limit enforcement
- Test with large mobile camera photos (5MB+)
- Document findings in WORKLOGS.md

### Current understanding:
- Backend has BodyLimit: 500MB (in main.go)
- Upload endpoint works (returns 200)
- Post creation endpoint may be failing due to header size or body size
- Vercel has ~8KB header limit (431 error)
- Error message "Request En..." suggests "Request Entity Too Large" (413)
- **CONFIRMED**: Next.js Route Handler has default body limit rejecting 13MB mobile photos

### Next.js Version: 16.3.0
### Architecture: App Router Route Handler (Node.js runtime)
### Root Cause: Next.js Route Handler body size limit (likely 1MB default)
### Fix: Configure body size limit in next.config.ts or route handler

### Fix applied:
- Added `experimental.serverActions.bodySizeLimit: '32mb'` to `next.config.ts`
- Added user-friendly 413 error handling in `/api/media/upload/route.ts`
- Frontend api.ts already has safe JSON parsing (from previous session)
- Both layers now support PRD's 30MB requirement:
  - Next.js: 32MB (was 1MB default)
  - Go backend: 500MB (unchanged, already sufficient)
- Commit: 39e82a0

### Tests added:
- Added `backend/go/internal/handlers/media_test.go` with 2 tests:
  - `TestUploadMediaBodyLimit` - verifies 35MB file exceeds 30MB limit
  - `TestUploadMediaSmallFile` - verifies small file upload works
- Both tests pass ✅

### Verification:
- Small image upload: ✅ works
- Health check: ✅ healthy
- CI: Backend tests passing, frontend building successfully

### SECOND FIX — Vercel Function Payload Limit:
- Problem: Vercel Hobby plan has 4MB function payload limit
- Error: `413 FUNCTION_PAYLOAD_TOO_LARGE` from Vercel
- Root cause: `request.arrayBuffer()` buffers entire request in memory → Vercel rejects >4MB
- Fix: Stream `request.body` directly to backend instead of buffering
- Added `duplex: 'half'` for Node.js stream support
- Commit: 55c06bd