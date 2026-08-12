'use client';

/**
 * Main app shell layout — wraps all authenticated pages.
 * Provides TopBar, Sidebar (desktop), BottomNav (mobile), and
 * an auth guard that redirects to /login if not authenticated.
 */

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { ROUTES } from '@/lib/constants';
import TopBar from '@/components/layout/TopBar';
import Sidebar from '@/components/layout/Sidebar';
import BottomNav from '@/components/layout/BottomNav';
import styles from './main.module.css';

export default function MainLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, user } = useAuth();
  const router = useRouter();

  // Auth guard: redirect to login if not authenticated (after loading completes)
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace(ROUTES.LOGIN);
    }
  }, [isLoading, isAuthenticated, router]);

  // Show a loading spinner while auth state hydrates
  if (isLoading) {
    return (
      <div className={styles.loadingScreen}>
        <div className={styles.loadingSpinner} />
      </div>
    );
  }

  // Don't render the shell if not authenticated (redirect is in progress)
  if (!isAuthenticated || !user) {
    return null;
  }

  return (
    <div className={styles.shell}>
      <TopBar />
      <div className={styles.body}>
        <Sidebar username={user.username} />
        <main className={styles.content}>{children}</main>
      </div>
      <BottomNav username={user.username} />
    </div>
  );
}
