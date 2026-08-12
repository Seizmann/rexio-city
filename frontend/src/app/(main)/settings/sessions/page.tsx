'use client';

/**
 * Active Sessions Page — /settings/sessions
 *
 * Displays all active login sessions (devices) for the current user.
 * Allows revoking individual sessions or logging out of all devices.
 * Calls GET /api/auth/sessions and POST /api/auth/sessions/:id/revoke.
 */

import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import type { Session } from '@/lib/types';
import { api } from '@/lib/api';
import { API, ROUTES } from '@/lib/constants';
import { useAuth } from '@/context/AuthContext';

/* ── Helpers ─────────────────────────────────────────────────── */

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
}

function parseDevice(userAgent: string): string {
  // Best-effort human-readable device description from User-Agent string
  if (!userAgent) return 'Unknown device';
  if (/iPhone|iPad/.test(userAgent)) return '📱 iOS Device';
  if (/Android/.test(userAgent)) return '📱 Android Device';
  if (/Mac OS/.test(userAgent)) return '💻 Mac';
  if (/Windows/.test(userAgent)) return '💻 Windows PC';
  if (/Linux/.test(userAgent)) return '🖥️ Linux';
  return '🌐 Browser';
}

function parseBrowser(userAgent: string): string {
  if (!userAgent) return '';
  if (/Firefox\//.test(userAgent)) return 'Firefox';
  if (/Edg\//.test(userAgent)) return 'Edge';
  if (/Chrome\//.test(userAgent)) return 'Chrome';
  if (/Safari\//.test(userAgent)) return 'Safari';
  return 'Unknown browser';
}

/* ── Component ───────────────────────────────────────────────── */

export default function SessionsPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading, logoutAll } = useAuth();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [revoking, setRevoking] = useState<number | null>(null);
  const [revokingAll, setRevokingAll] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.replace(ROUTES.LOGIN);
    }
  }, [authLoading, isAuthenticated, router]);

  const fetchSessions = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.get<Session[]>(API.AUTH_SESSIONS);
      if (res.success && res.data) {
        setSessions(res.data);
      } else {
        setError(res.error?.message ?? 'Failed to load sessions');
      }
    } catch {
      setError('Network error — please try again');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (isAuthenticated) void fetchSessions();
  }, [isAuthenticated, fetchSessions]);

  const revokeSession = async (sessionId: number) => {
    setRevoking(sessionId);
    try {
      const res = await api.post(API.AUTH_SESSION_REVOKE(sessionId));
      if (res.success) {
        setSessions((prev) => prev.filter((s) => s.id !== sessionId));
      } else {
        setError('Failed to revoke session');
      }
    } catch {
      setError('Network error');
    } finally {
      setRevoking(null);
    }
  };

  const handleLogoutAll = async () => {
    setRevokingAll(true);
    try {
      await logoutAll();
      // logoutAll clears auth state; router will redirect via the useEffect above
    } catch {
      setError('Failed to log out of all devices');
      setRevokingAll(false);
    }
  };

  if (authLoading || (!isAuthenticated && !authLoading)) {
    return null; // redirect in progress
  }

  return (
    <div className="sessions-page">
      <div className="sessions-container">
        {/* Header */}
        <div className="sessions-header">
          <button className="back-btn" onClick={() => router.back()} aria-label="Go back">
            ← Back
          </button>
          <h1 className="sessions-title">Active Sessions</h1>
          <p className="sessions-subtitle">
            These devices are currently signed into your account. Revoke any session you don&apos;t recognise.
          </p>
        </div>

        {/* Error */}
        {error && (
          <div className="sessions-error" role="alert">
            {error}
            <button onClick={() => setError(null)} aria-label="Dismiss">✕</button>
          </div>
        )}

        {/* Sessions List */}
        {loading ? (
          <div className="sessions-loading">
            {[1, 2, 3].map((i) => (
              <div key={i} className="session-skeleton" />
            ))}
          </div>
        ) : sessions.length === 0 ? (
          <div className="sessions-empty">No active sessions found.</div>
        ) : (
          <ul className="sessions-list" role="list">
            {sessions.map((session) => (
              <li key={session.id} className="session-card">
                <div className="session-icon" aria-hidden="true">
                  {parseDevice(session.device_info).slice(0, 2)}
                </div>
                <div className="session-info">
                  <p className="session-device">
                    {parseDevice(session.device_info)} — {parseBrowser(session.device_info)}
                  </p>
                  <p className="session-ip">{session.ip_address || 'Unknown IP'}</p>
                  <p className="session-dates">
                    Signed in {formatDate(session.created_at)}
                    {session.last_used_at !== session.created_at && (
                      <> · Last active {formatDate(session.last_used_at)}</>
                    )}
                  </p>
                </div>
                <button
                  id={`revoke-session-${session.id}`}
                  className="revoke-btn"
                  onClick={() => void revokeSession(session.id)}
                  disabled={revoking === session.id}
                  aria-label={`Revoke session from ${parseDevice(session.device_info)}`}
                >
                  {revoking === session.id ? 'Revoking…' : 'Revoke'}
                </button>
              </li>
            ))}
          </ul>
        )}

        {/* Logout All */}
        {sessions.length > 0 && (
          <div className="sessions-footer">
            <button
              id="logout-all-devices-btn"
              className="logout-all-btn"
              onClick={() => void handleLogoutAll()}
              disabled={revokingAll}
              aria-label="Log out of all devices"
            >
              {revokingAll ? 'Logging out everywhere…' : '🚪 Log out of all devices'}
            </button>
            <p className="logout-all-note">
              This will end every active session including this one.
            </p>
          </div>
        )}
      </div>

      <style>{`
        .sessions-page {
          min-height: 100vh;
          background: var(--bg-primary);
          padding: 1.5rem 1rem;
        }
        .sessions-container {
          max-width: 680px;
          margin: 0 auto;
        }
        .sessions-header {
          margin-bottom: 2rem;
        }
        .back-btn {
          background: none;
          border: none;
          color: var(--accent);
          font-size: 0.9rem;
          cursor: pointer;
          padding: 0;
          margin-bottom: 1rem;
        }
        .back-btn:hover { opacity: 0.8; }
        .sessions-title {
          font-size: 1.5rem;
          font-weight: 700;
          color: var(--text-primary);
          margin: 0 0 0.5rem;
        }
        .sessions-subtitle {
          color: var(--text-secondary);
          font-size: 0.9rem;
          margin: 0;
        }
        .sessions-error {
          background: color-mix(in srgb, var(--error, #e53e3e) 12%, transparent);
          border: 1px solid var(--error, #e53e3e);
          border-radius: 8px;
          padding: 0.75rem 1rem;
          color: var(--error, #e53e3e);
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 1.5rem;
          font-size: 0.875rem;
        }
        .sessions-error button {
          background: none;
          border: none;
          cursor: pointer;
          color: inherit;
          font-size: 1rem;
          padding: 0;
        }
        .sessions-loading {
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
        }
        .session-skeleton {
          height: 88px;
          border-radius: 12px;
          background: var(--bg-secondary);
          animation: skeleton-pulse 1.5s ease-in-out infinite;
        }
        @keyframes skeleton-pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
        .sessions-empty {
          text-align: center;
          color: var(--text-secondary);
          padding: 3rem 0;
        }
        .sessions-list {
          list-style: none;
          padding: 0;
          margin: 0;
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
        }
        .session-card {
          display: flex;
          align-items: flex-start;
          gap: 1rem;
          padding: 1rem 1.25rem;
          background: var(--bg-secondary);
          border: 1px solid var(--border);
          border-radius: 12px;
          transition: border-color 0.2s;
        }
        .session-card:hover { border-color: var(--accent); }
        .session-icon {
          font-size: 1.5rem;
          flex-shrink: 0;
          padding-top: 0.1rem;
        }
        .session-info {
          flex: 1;
          min-width: 0;
        }
        .session-device {
          font-weight: 600;
          color: var(--text-primary);
          margin: 0 0 0.25rem;
          font-size: 0.95rem;
        }
        .session-ip {
          font-size: 0.8rem;
          color: var(--text-secondary);
          margin: 0 0 0.25rem;
          font-family: monospace;
        }
        .session-dates {
          font-size: 0.8rem;
          color: var(--text-muted, var(--text-secondary));
          margin: 0;
        }
        .revoke-btn {
          flex-shrink: 0;
          background: none;
          border: 1px solid color-mix(in srgb, var(--error, #e53e3e) 60%, transparent);
          color: var(--error, #e53e3e);
          border-radius: 8px;
          padding: 0.4rem 0.85rem;
          font-size: 0.8rem;
          font-weight: 600;
          cursor: pointer;
          transition: background 0.2s;
        }
        .revoke-btn:hover:not(:disabled) {
          background: color-mix(in srgb, var(--error, #e53e3e) 10%, transparent);
        }
        .revoke-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }
        .sessions-footer {
          margin-top: 2rem;
          padding-top: 1.5rem;
          border-top: 1px solid var(--border);
          text-align: center;
        }
        .logout-all-btn {
          background: color-mix(in srgb, var(--error, #e53e3e) 12%, transparent);
          border: 1px solid color-mix(in srgb, var(--error, #e53e3e) 50%, transparent);
          color: var(--error, #e53e3e);
          border-radius: 10px;
          padding: 0.75rem 1.5rem;
          font-size: 0.9rem;
          font-weight: 600;
          cursor: pointer;
          transition: background 0.2s;
        }
        .logout-all-btn:hover:not(:disabled) {
          background: color-mix(in srgb, var(--error, #e53e3e) 20%, transparent);
        }
        .logout-all-btn:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }
        .logout-all-note {
          font-size: 0.8rem;
          color: var(--text-secondary);
          margin: 0.5rem 0 0;
        }
      `}</style>
    </div>
  );
}
