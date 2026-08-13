/**
 * Custom hook for fetching data with caching support.
 *
 * Provides:
 * - Immediate display of cached data on repeat visits
 * - Background revalidation to keep data fresh
 * - Loading state only on initial fetch (not on cache hits)
 *
 * Usage:
 * ```tsx
 * const { data, loading, error } = useCachedFetch(
 *   'feed-following',
 *   () => api.get<Post[]>('/api/feed?tab=following&page=1'),
 * );
 * ```
 */

import { useState, useEffect, useRef } from 'react';
import { fetchWithCache, getCached } from './cache';

interface FetchState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

interface UseCachedFetchOptions<T> {
  /** Re-fetch when this value changes (e.g., tab change, user change) */
  dependencies?: unknown[];
  /** Force re-fetch even if cache exists */
  forceRefresh?: boolean;
  /** Callback when data changes */
  onSuccess?: (data: T) => void;
}

export function useCachedFetch<T>(
  key: string,
  fetcher: () => Promise<T>,
  options: UseCachedFetchOptions<T> = {},
): FetchState<T> {
  const {
    dependencies = [],
    forceRefresh = false,
    onSuccess,
  } = options;

  const [state, setState] = useState<FetchState<T>>({
    data: null,
    loading: true,
    error: null,
  });

  const isSubscribedRef = useRef(true);
  const fetchPromiseRef = useRef<Promise<void> | null>(null);

  // Single effect that handles both initial load and cache hits
  useEffect(() => {
    isSubscribedRef.current = true;

    // Check cache first for immediate display
    const cachedData = getCached<T>(key);

    if (cachedData !== null && !forceRefresh) {
      // Use setTimeout to defer state update and avoid cascading renders
      const timer = setTimeout(() => {
        setState({ data: cachedData, loading: false, error: null });

        // Re-fetch in background to update cache (fire and forget)
        fetchPromiseRef.current = fetchWithCache(key, fetcher).then((freshData) => {
          if (isSubscribedRef.current) {
            setState({ data: freshData, loading: false, error: null });
            onSuccess?.(freshData);
          }
        }).catch(() => {
          // Ignore background revalidation errors
        });
      }, 0);

      return () => {
        clearTimeout(timer);
        isSubscribedRef.current = false;
        fetchPromiseRef.current = null;
      };
    }

    // No cache or force refresh - fetch fresh
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setState({ data: null, loading: true, error: null });
    fetchPromiseRef.current = fetchWithCache(key, fetcher)
      .then((data) => {
        if (isSubscribedRef.current) {
          setState({ data, loading: false, error: null });
          onSuccess?.(data);
        }
      })
      .catch((err) => {
        if (isSubscribedRef.current) {
          const errorMessage =
            err instanceof Error ? err.message : 'Failed to fetch data';
          setState({ data: null, loading: false, error: errorMessage });
        }
      });

    return () => {
      isSubscribedRef.current = false;
      fetchPromiseRef.current = null;
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, ...dependencies]);

  return state;
}
