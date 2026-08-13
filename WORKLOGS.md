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

### Files modified:
- `src/lib/cache.ts` (new)
- `src/lib/useCachedFetch.ts` (new)
- `src/app/(main)/page.tsx` (modified)
- `src/app/(main)/[username]/page.tsx` (modified)

### Notes for next agent:
- Messages and Notifications pages don't exist yet (routes `/messages` and `/notifications` not implemented)
- When those pages are created, apply the same `useCachedFetch` pattern
- Cache TTL is 5 minutes — can be adjusted in `cache.ts` if needed
