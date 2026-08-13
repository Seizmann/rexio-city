/**
 * Simple in-memory cache for API responses.
 *
 * Provides stale-while-revalidate behavior:
 * 1. Return cached data immediately if available
 * 2. Re-fetch in background to update cache
 * 3. Deduplicate concurrent requests for the same key
 *
 * This eliminates the loading delay on repeat navigation visits
 * while still keeping data fresh.
 */

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  promise?: Promise<T>;
}

// In-memory cache store
const cache = new Map<string, CacheEntry<unknown>>();

// TTL for cached data (5 minutes)
const CACHE_TTL_MS = 5 * 60 * 1000;

/**
 * Get cached data if available and not expired.
 */
export function getCached<T>(key: string): T | null {
  const entry = cache.get(key) as CacheEntry<T> | undefined;
  if (!entry) return null;

  // Check if expired
  if (Date.now() - entry.timestamp > CACHE_TTL_MS) {
    cache.delete(key);
    return null;
  }

  return entry.data;
}

/**
 * Set cache entry.
 */
export function setCached<T>(key: string, data: T): void {
  cache.set(key, {
    data,
    timestamp: Date.now(),
  });
}

/**
 * Remove cache entry.
 */
export function invalidateCache(key: string): void {
  cache.delete(key);
}

/**
 * Clear entire cache.
 */
export function clearCache(): void {
  cache.clear();
}

/**
 * Fetch data with caching.
 *
 * - Returns cached data immediately if available
 * - Re-fetches in background if cache is stale or absent
 * - Deduplicates concurrent requests for the same key
 */
export async function fetchWithCache<T>(
  key: string,
  fetcher: () => Promise<T>,
  options?: {
    /** Force re-fetch even if cache exists */
    forceRefresh?: boolean;
  },
): Promise<T> {
  const { forceRefresh = false } = options || {};

  // Check cache first
  if (!forceRefresh) {
    const cached = getCached<T>(key);
    if (cached !== null) {
      // Re-fetch in background to update cache
      setTimeout(() => {
        void fetcher().then((freshData) => setCached(key, freshData)).catch(() => {
          // Ignore errors during background revalidation
        });
      }, 0);
      return cached;
    }
  }

  // Check if there's already a pending request
  const existing = cache.get(key) as CacheEntry<T> | undefined;
  if (existing?.promise && !forceRefresh) {
    return existing.promise;
  }

  // Create new fetch promise
  const promise = fetcher().then((data) => {
    setCached(key, data);
    return data;
  });

  cache.set(key, {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
    data: undefined as unknown as T,
    timestamp: Date.now(),
    promise,
  });

  return promise;
}
