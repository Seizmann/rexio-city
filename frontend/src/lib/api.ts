/**
 * Central API client for RexiO City frontend.
 *
 * Architecture (per AGENTS.md D1 & DECISIONS.md):
 *   Browser → fetch('/api/...') → Next.js middleware proxy → Go backend
 *
 * Token storage uses localStorage for MVP. Access tokens are short-lived (15m),
 * refresh tokens are long-lived (30d). On 401, the client attempts a silent
 * refresh before failing.
 */

import type { APIResponse, AuthData, LoginPayload, SignupPayload } from './types';
import { API } from './constants';

/* ── Token Storage ────────────────────────────────────────────── */

const TOKEN_KEY = 'rexio_access_token';
const REFRESH_KEY = 'rexio_refresh_token';
const USER_KEY = 'rexio_user';

export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function getRefreshToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(REFRESH_KEY);
}

export function setTokens(access: string, refresh: string): void {
  localStorage.setItem(TOKEN_KEY, access);
  localStorage.setItem(REFRESH_KEY, refresh);
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem(USER_KEY);
}

export function setStoredUser(user: unknown): void {
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

/* ── Core Fetch Wrapper ───────────────────────────────────────── */

/** Flag to prevent concurrent refresh attempts. */
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

/**
 * Low-level fetch wrapper. Attaches auth header, parses JSON response,
 * and attempts a single token refresh on 401.
 */
async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
  skipAuth = false,
): Promise<APIResponse<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  // Attach auth header unless explicitly skipped (login/signup/refresh)
  if (!skipAuth) {
    const token = getAccessToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  const response = await fetch(path, {
    ...options,
    headers,
  });

  // Attempt silent refresh on 401
  if (response.status === 401 && !skipAuth) {
    const refreshed = await attemptTokenRefresh();
    if (refreshed) {
      // Retry original request with new token
      const retryHeaders = { ...headers };
      const newToken = getAccessToken();
      if (newToken) {
        retryHeaders['Authorization'] = `Bearer ${newToken}`;
      }
      const retryResponse = await fetch(path, { ...options, headers: retryHeaders });
      return retryResponse.json() as Promise<APIResponse<T>>;
    }
    // Refresh failed — clear tokens and let the caller handle it
    clearTokens();
  }

  return response.json() as Promise<APIResponse<T>>;
}

/**
 * Attempt to refresh the access token using the stored refresh token.
 * Deduplicates concurrent refresh attempts.
 */
async function attemptTokenRefresh(): Promise<boolean> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const response = await fetch(API.AUTH_REFRESH, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) return false;

      const data = (await response.json()) as APIResponse<AuthData>;
      if (data.success && data.data) {
        setTokens(data.data.access_token, data.data.refresh_token);
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

/* ── Auth API Calls (no auth header needed) ───────────────────── */

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

/* ── Generic Authenticated API Methods ────────────────────────── */

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
};
