'use client';

/**
 * Toast notification system for RexiO City.
 * Provides a ToastProvider context and useToast() hook for
 * showing success/error messages from anywhere in the app.
 */

import {
  createContext,
  useContext,
  useState,
  useCallback,
  type ReactNode,
} from 'react';
import styles from './Toast.module.css';

/* ── Types ────────────────────────────────────────────────────── */

type ToastType = 'info' | 'success' | 'error';

interface Toast {
  id: number;
  message: string;
  type: ToastType;
}

interface ToastContextValue {
  showToast: (message: string, type?: ToastType) => void;
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

let nextId = 0;

/* ── Provider ─────────────────────────────────────────────────── */

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const showToast = useCallback((message: string, type: ToastType = 'info') => {
    const id = nextId++;
    setToasts((prev) => [...prev, { id, message, type }]);

    // Auto-dismiss after 4 seconds
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4000);
  }, []);

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      {toasts.length > 0 && (
        <div className={styles.toastContainer} role="status" aria-live="polite">
          {toasts.map((toast) => (
            <div
              key={toast.id}
              className={`${styles.toast} ${toast.type === 'error' ? styles.toastError : ''} ${toast.type === 'success' ? styles.toastSuccess : ''}`}
            >
              <span>{toast.message}</span>
              <button
                className={styles.closeBtn}
                onClick={() => dismiss(toast.id)}
                aria-label="Dismiss"
              >
                ✕
              </button>
            </div>
          ))}
        </div>
      )}
    </ToastContext.Provider>
  );
}

/* ── Hook ─────────────────────────────────────────────────────── */

export function useToast(): ToastContextValue {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
}
