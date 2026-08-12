'use client';

/**
 * Auth context for RexiO City.
 * Manages JWT token lifecycle, user state, and provides login/signup/logout
 * functions to the entire app via React Context.
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
  apiLogin,
  apiSignup,
  api,
  getAccessToken,
  setTokens,
  clearTokens,
  setStoredUser,
  getStoredUser,
} from '@/lib/api';
import { API } from '@/lib/constants';

/* ── Context Types ────────────────────────────────────────────── */

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>;
  signup: (data: {
    username: string;
    email: string;
    password: string;
    display_name: string;
  }) => Promise<void>;
  logout: () => void;
  /** Update the cached user object (e.g. after profile edit). */
  setUser: (user: User) => void;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

/* ── Provider ─────────────────────────────────────────────────── */

export function AuthProvider({ children }: { children: ReactNode }) {
  // Initialize state directly from storage to avoid synchronous setState inside useEffect
  const [state, setState] = useState<AuthState>(() => {
    if (typeof window === 'undefined') {
      return { user: null, isLoading: true, isAuthenticated: false };
    }
    const token = getAccessToken();
    if (!token) {
      return { user: null, isLoading: false, isAuthenticated: false };
    }
    const cached = getStoredUser() as User | null;
    return { user: cached, isLoading: true, isAuthenticated: !!cached };
  });

  // Verify auth token with backend on mount
  useEffect(() => {
    const token = getAccessToken();
    if (!token) return;

    api
      .get<User>(API.USERS_ME)
      .then((res) => {
        if (res.success && res.data) {
          setStoredUser(res.data);
          setState({ user: res.data, isLoading: false, isAuthenticated: true });
        } else {
          clearTokens();
          setState({ user: null, isLoading: false, isAuthenticated: false });
        }
      })
      .catch(() => {
        clearTokens();
        setState({ user: null, isLoading: false, isAuthenticated: false });
      });
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiLogin({ email, password });
    if (res.success && res.data) {
      setTokens(res.data.access_token, res.data.refresh_token);
      setStoredUser(res.data.user);
      setState({
        user: res.data.user,
        isLoading: false,
        isAuthenticated: true,
      });
    } else {
      throw new Error(res.error?.message || 'Login failed');
    }
  }, []);

  const signup = useCallback(
    async (data: {
      username: string;
      email: string;
      password: string;
      display_name: string;
    }) => {
      const res = await apiSignup(data);
      if (res.success && res.data) {
        setTokens(res.data.access_token, res.data.refresh_token);
        setStoredUser(res.data.user);
        setState({
          user: res.data.user,
          isLoading: false,
          isAuthenticated: true,
        });
      } else {
        throw new Error(res.error?.message || 'Signup failed');
      }
    },
    [],
  );

  const logout = useCallback(() => {
    clearTokens();
    setState({ user: null, isLoading: false, isAuthenticated: false });
  }, []);

  const setUser = useCallback((user: User) => {
    setStoredUser(user);
    setState((prev) => ({ ...prev, user }));
  }, []);

  return (
    <AuthContext.Provider
      value={{ ...state, login, signup, logout, setUser }}
    >
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
