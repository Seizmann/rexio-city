'use client';

/**
 * Client-side providers wrapper.
 * Combines all context providers (Auth, Toast, etc.) into a single
 * component so the root layout stays clean.
 */

import { type ReactNode } from 'react';
import { AuthProvider } from '@/context/AuthContext';
import { ToastProvider } from '@/components/ui/Toast';

export default function Providers({ children }: { children: ReactNode }) {
  return (
    <AuthProvider>
      <ToastProvider>{children}</ToastProvider>
    </AuthProvider>
  );
}
