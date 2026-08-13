'use client';

/**
 * Main app shell layout — wraps pages in the (main) group.
 *
 * For authenticated users: renders TopBar, Sidebar, content area, and BottomNav.
 * For unauthenticated users on home page (/): renders the page directly (LandingAuth screen).
 * For unauthenticated users on other routes: redirects to /login.
 */

import { useRouter, usePathname } from 'next/navigation';
import { useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { ROUTES } from '@/lib/constants';
import TopBar from '@/components/layout/TopBar';
import Sidebar from '@/components/layout/Sidebar';
import BottomNav from '@/components/layout/BottomNav';
import SplashScreen from '@/components/ui/SplashScreen';
import DebugPanel from '@/components/debug/DebugPanel';
import styles from './main.module.css';

export default function MainLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, user } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  const isHomePage = pathname === '/';

  // Auth guard: redirect to login if not authenticated and trying to access protected routes
  useEffect(() => {
    if (!isLoading && !isAuthenticated && !isHomePage) {
      router.replace(ROUTES.LOGIN);
    }
  }, [isLoading, isAuthenticated, isHomePage, router]);

  // Show splash screen while auth hydrates (Twitter-style icon animation)
  if (isLoading) {
    return <SplashScreen />;
  }

  // If not authenticated and on home page (/), render page content directly (LandingAuth screen)
  if (!isAuthenticated) {
    if (isHomePage) {
      return <div suppressHydrationWarning>{children}</div>;
    }
    return null;
  }

  // Authenticated user: render full App Shell
  return (
    <div className={styles.shell} suppressHydrationWarning>
      <TopBar />
      <div className={styles.body} suppressHydrationWarning>
        <Sidebar username={user?.username || ''} />
        <main className={styles.content} suppressHydrationWarning>{children}</main>
      </div>
      <BottomNav username={user?.username || ''} />
      <DebugPanel />
    </div>
  );
}
