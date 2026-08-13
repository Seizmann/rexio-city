/**
 * Central API client for RexiO City frontend.
 *
 * Architecture (per AGENTS.md D1 & DECISIONS.md):
 *   Browser → fetch('/api/...') → Next.js middleware proxy → Go backend
 *
 * Token storage (post-security-migration):
 *   - access_token: held ONLY in module-level memory (never localStorage/cookie).
 *   - refresh_token: NEVER accessible to JS. Stored as httpOnly cookie by backend.
 *   - user profile: cached in localStorage (not sensitive; no tokens).
 *   - CSRF token: read from the readable "rexio_csrf" cookie, sent as
 *     X-CSRF-Token header on all state-changing requests.
 *
 * On app load / hard refresh, the module calls attemptTokenRefresh() immediately
 * (see AuthContext.tsx) to re-obtain an access token from the httpOnly-cookie-based
 * refresh endpoint. This is transparent to the user.
 */

import type { APIResponse, AuthData, LoginPayload, SignupPayload } from './types';
import { API } from './constants';

/* ── In-Memory Access Token Store ──────────────────────────────── */
// This is the ONLY place the access token ever lives in the browser.
// It is lost on page refresh — that's intentional. AuthContext handles
// silent re-hydration via the refresh endpoint on mount.

let _accessToken: string | null = null;

export function getAccessToken(): string | null {
  return _accessToken;
}

export function setAccessToken(token: string | null): void {
  _accessToken = token;
}

/* ── Non-Sensitive User Profile Cache (localStorage) ───────────── */
// The user object (username, avatar, etc.) is not sensitive — it's
// publicly visible data. We cache it to avoid a flash of "loading"
// on hard refresh while the silent refresh is in flight.

const USER_KEY = 'rexio_user';

export function setStoredUser(user: unknown): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function getStoredUser(): unknown {
  if (typeof window === 'undefined') return null;
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function clearStoredUser(): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(USER_KEY);
}

/* ── CSRF Token (double-submit cookie pattern) ─────────────────── */
// The backend sets a readable (non-httpOnly) cookie "rexio_csrf".
// We read it here and send it as X-CSRF-Token on state-changing requests.
// An attacker cannot read cookies from a different origin, so they cannot
// forge this header even if they trigger the cookie automatically.

function getCSRFToken(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)rexio_csrf=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

/* ── Core Fetch Wrapper ─────────────────────────────────────────── */

/** Flag to prevent concurrent silent-refresh attempts. */
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

/**
 * Safely parse JSON response. Handles non-JSON responses (e.g., 413, 431, 500)
 * that Vercel/Cloudflare may return as plain text instead of JSON.
 */
async function parseJSONResponse<T>(response: Response): Promise<APIResponse<T>> {
  const contentType = response.headers.get('content-type') || '';

  // If response is not JSON, read as text and throw a user-friendly error
  if (!contentType.includes('application/json')) {
    const text = await response.text();
    const errorMessage = text.includes('Entity')
      ? 'File too large. Please compress your photo and try again.'
      : text.includes('Header')
        ? 'Request headers too large. Please refresh the page and try again.'
        : text || `Server error (${response.status})`;
    throw new Error(errorMessage);
  }

  try {
    return await response.json() as APIResponse<T>;
  } catch {
    throw new Error(`Invalid JSON response from server (${response.status})`);
  }
}

/**
 * Low-level fetch wrapper. Attaches auth Bearer header, attaches CSRF token
 * on mutating requests, parses JSON response, and attempts a single silent
 * token refresh on 401 before giving up.
 */
async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
  skipAuth = false,
): Promise<APIResponse<T>> {
  const method = (options.method ?? 'GET').toUpperCase();
  const isMutating = method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS';

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (!skipAuth && _accessToken) {
    headers['Authorization'] = `Bearer ${_accessToken}`;
  }

  // Attach CSRF token on all state-changing requests (POST/PUT/PATCH/DELETE)
  if (isMutating && !skipAuth) {
    const csrf = getCSRFToken();
    if (csrf) {
      headers['X-CSRF-Token'] = csrf;
    }
  }

  const response = await fetch(path, {
    ...options,
    headers,
    // Credentials=include ensures the httpOnly refresh cookie and CSRF cookie
    // are sent with requests to the same origin (Next.js proxy).
    credentials: 'include',
  });

  // Check for 431 (header too large) — Vercel returns plain text, not JSON
  if (response.status === 431) {
    const text = await response.text();
    throw new Error(text.includes('Request Header') ? 'Request headers too large. Please clear cookies and try again.' : text);
  }

  // Attempt silent refresh on 401
  if (response.status === 401 && !skipAuth) {
    const refreshed = await attemptTokenRefresh();
    if (refreshed) {
      // Retry original request with new in-memory access token
      const retryHeaders = { ...headers };
      if (_accessToken) {
        retryHeaders['Authorization'] = `Bearer ${_accessToken}`;
      }
      // Re-read CSRF after refresh (backend rotates the CSRF cookie too)
      if (isMutating) {
        const csrf = getCSRFToken();
        if (csrf) retryHeaders['X-CSRF-Token'] = csrf;
      }
      const retryResponse = await fetch(path, {
        ...options,
        headers: retryHeaders,
        credentials: 'include',
      });
      return parseJSONResponse<T>(retryResponse);
    }
    // Refresh failed — clear in-memory token and cached user
    setAccessToken(null);
    clearStoredUser();
  }

  return parseJSONResponse<T>(response);
}

/**
 * Attempt a silent token refresh using the httpOnly refresh cookie.
 * The browser sends the cookie automatically — no manual cookie handling.
 * Deduplicates concurrent refresh attempts to avoid race conditions.
 */
export async function attemptTokenRefresh(): Promise<boolean> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      // POST to refresh endpoint — browser sends httpOnly cookie automatically.
      // No body needed; backend reads the cookie.
      const response = await fetch(API.AUTH_REFRESH, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include', // sends the httpOnly refresh cookie
      });

      if (!response.ok) return false;

      const data = (await response.json()) as APIResponse<AuthData>;
      if (data.success && data.data?.access_token) {
        // Store only the access token in memory, never anywhere else
        setAccessToken(data.data.access_token);
        return true;
      }
      return false;
    } catch {
      return false;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

/* ── Auth API Calls ─────────────────────────────────────────────── */

export async function apiLogin(payload: LoginPayload): Promise<APIResponse<AuthData>> {
  return apiFetch<AuthData>(API.AUTH_LOGIN, {
    method: 'POST',
    body: JSON.stringify(payload),
  }, true);
}

export async function apiSignup(payload: SignupPayload): Promise<APIResponse<AuthData>> {
  return apiFetch<AuthData>(API.AUTH_SIGNUP, {
    method: 'POST',
    body: JSON.stringify(payload),
  }, true);
}

export async function apiLogout(): Promise<void> {
  try {
    await apiFetch(API.AUTH_LOGOUT, { method: 'POST' });
  } finally {
    // Clear in-memory state regardless of server response
    setAccessToken(null);
    clearStoredUser();
  }
}

export async function apiLogoutAll(): Promise<void> {
  try {
    await apiFetch(API.AUTH_LOGOUT_ALL, { method: 'POST' });
  } finally {
    setAccessToken(null);
    clearStoredUser();
  }
}

/* ── Generic Authenticated API Methods ──────────────────────────── */

export const api = {
  get: <T>(path: string) =>
    apiFetch<T>(path, { method: 'GET' }),

  post: <T>(path: string, body?: unknown) =>
    apiFetch<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),

  patch: <T>(path: string, body?: unknown) =>
    apiFetch<T>(path, {
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(path: string, body?: unknown) =>
    apiFetch<T>(path, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(path: string) =>
    apiFetch<T>(path, { method: 'DELETE' }),

  upload: async <T>(path: string, formData: FormData): Promise<APIResponse<T>> => {
    const headers: Record<string, string> = {};
    const token = getAccessToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const csrf = getCSRFToken();
    if (csrf) {
      headers['X-CSRF-Token'] = csrf;
    }

    // Log the request for debugging
    console.log('[api.upload] Starting upload:', path, 'hasToken:', !!token, 'hasCSRF:', !!csrf);
    console.log('[api.upload] Headers:', JSON.stringify(headers));

    let response = await fetch(path, {
      method: 'POST',
      headers,
      body: formData,
      credentials: 'include',
    });

    console.log('[api.upload] Response status:', response.status, 'content-type:', response.headers.get('content-type'));

    // Check for 431 (header too large) — Vercel returns plain text, not JSON
    if (response.status === 431) {
      const text = await response.text();
      console.error('[api.upload] 431 Response:', text.slice(0, 200));
      throw new Error(text.includes('Request Header') ? 'Request headers too large. Please clear cookies and try again.' : text);
    }

    if (response.status === 401) {
      console.log('[api.upload] 401 received, attempting token refresh');
      const refreshed = await attemptTokenRefresh();
      if (refreshed) {
        console.log('[api.upload] Token refresh succeeded, retrying');
        const retryHeaders: Record<string, string> = {};
        const newToken = getAccessToken();
        if (newToken) retryHeaders['Authorization'] = `Bearer ${newToken}`;
        const newCsrf = getCSRFToken();
        if (newCsrf) retryHeaders['X-CSRF-Token'] = newCsrf;

        response = await fetch(path, {
          method: 'POST',
          headers: retryHeaders,
          body: formData,
          credentials: 'include',
        });
        console.log('[api.upload] Retry response status:', response.status);

        // Check for 431 on retry too
        if (response.status === 431) {
          const text = await response.text();
          throw new Error(text.includes('Request Header') ? 'Request headers too large. Please clear cookies and try again.' : text);
        }
      } else {
        console.error('[api.upload] Token refresh failed');
      }
    }

    return parseJSONResponse<T>(response);
  },
};
