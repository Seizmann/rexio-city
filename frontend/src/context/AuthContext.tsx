'use client';

/**
 * Auth context for RexiO City.
 *
 * Post-security-migration design:
 *   - access_token lives ONLY in the api.ts module memory (_accessToken variable).
 *   - refresh_token is NEVER accessible here or anywhere in JS (httpOnly cookie).
 *   - On mount, we call attemptTokenRefresh() silently; if it succeeds we have a
 *     valid access token in memory and can fetch the user profile.
 *   - User profile is cached in localStorage (non-sensitive, safe to store).
 *   - logout() calls the server to revoke the session + clears local state.
 */

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react';
import type { User } from '@/lib/types';
import {
  attemptTokenRefresh,
  setAccessToken,
  api,
  apiLogin,
  apiSignup,
  apiLogout,
  apiLogoutAll,
  setStoredUser,
  getStoredUser,
  clearStoredUser,
} from '@/lib/api';
import { API } from '@/lib/constants';

/* ── Context Types ────────────────────────────────────────────── */

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

interface AuthContextValue extends AuthState {
  login:   (email: string, password: string) => Promise<void>;
  signup:  (data: { username: string; email: string; password: string; display_name: string }) => Promise<void>;
  logout:  () => Promise<void>;
  logoutAll: () => Promise<void>;
  /** Update the cached user object (e.g. after profile edit). */
  setUser: (user: User) => void;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

/* ── Provider ─────────────────────────────────────────────────── */

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    isLoading: true,
    isAuthenticated: false,
  });

  /**
   * On mount: attempt a silent refresh to re-hydrate the access token from
   * the httpOnly cookie. Show cached user immediately to avoid flash of
   * "not logged in" while the refresh is in flight.
   */
  useEffect(() => {
    const cached = getStoredUser() as User | null;

    void attemptTokenRefresh().then((refreshed) => {
      if (!refreshed) {
        // No valid refresh cookie — user is logged out
        clearStoredUser();
        setState({ user: null, isLoading: false, isAuthenticated: false });
        return;
      }

      // Refresh succeeded: fetch up-to-date profile
      void api
        .get<User>(API.USERS_ME)
        .then((res) => {
          if (res.success && res.data) {
            setStoredUser(res.data);
            setState({ user: res.data, isLoading: false, isAuthenticated: true });
          } else {
            clearStoredUser();
            setState({ user: null, isLoading: false, isAuthenticated: false });
          }
        })
        .catch(() => {
          // Network error — trust cached user if we have one
          if (cached) {
            setState({ user: cached, isLoading: false, isAuthenticated: true });
          } else {
            setState({ user: null, isLoading: false, isAuthenticated: false });
          }
        });
    });
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiLogin({ email, password });
    if (res.success && res.data) {
      // access_token goes into memory via api.ts; refresh token is set as
      // httpOnly cookie by the backend — we never see it here.
      setAccessToken(res.data.access_token);
      setStoredUser(res.data.user);
      setState({ user: res.data.user, isLoading: false, isAuthenticated: true });
    } else {
      throw new Error(res.error?.message ?? 'Login failed');
    }
  }, []);

  const signup = useCallback(async (data: {
    username: string;
    email: string;
    password: string;
    display_name: string;
  }) => {
    const res = await apiSignup(data);
    if (res.success && res.data) {
      setAccessToken(res.data.access_token);
      setStoredUser(res.data.user);
      setState({ user: res.data.user, isLoading: false, isAuthenticated: true });
    } else {
      throw new Error(res.error?.message ?? 'Signup failed');
    }
  }, []);

  const logout = useCallback(async () => {
    // Server revokes the session and clears the httpOnly cookie.
    // apiLogout() also clears the in-memory token and cached user.
    await apiLogout();
    setState({ user: null, isLoading: false, isAuthenticated: false });
  }, []);

  const logoutAll = useCallback(async () => {
    await apiLogoutAll();
    setState({ user: null, isLoading: false, isAuthenticated: false });
  }, []);

  const setUser = useCallback((user: User) => {
    setStoredUser(user);
    setState((prev) => ({ ...prev, user }));
  }, []);

  return (
    <AuthContext.Provider value={{ ...state, login, signup, logout, logoutAll, setUser }}>
      {children}
    </AuthContext.Provider>
  );
}

/* ── Hook ─────────────────────────────────────────────────────── */

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
