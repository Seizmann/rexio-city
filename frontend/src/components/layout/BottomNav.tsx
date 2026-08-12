'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ROUTES } from '@/lib/constants';
import styles from './BottomNav.module.css';

/** SVG icon paths — inline to avoid extra network requests for a few small icons. */
const icons = {
  home: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={styles.navIcon}>
      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  ),
  messages: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={styles.navIcon}>
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
    </svg>
  ),
  notifications: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={styles.navIcon}>
      <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
      <path d="M13.73 21a2 2 0 0 1-3.46 0" />
    </svg>
  ),
  profile: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={styles.navIcon}>
      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
      <circle cx="12" cy="7" r="4" />
    </svg>
  ),
};

interface NavItem {
  href: string;
  label: string;
  icon: React.ReactNode;
}

/**
 * Bottom tab bar for mobile navigation.
 * DESIGN.md §4: "Feed, Messaging, Notifications/Alerts" with icons + labels.
 * Hidden on desktop (≥768px) where Sidebar is used instead.
 */
export default function BottomNav({ username }: { username: string }) {
  const pathname = usePathname();

  const items: NavItem[] = [
    { href: ROUTES.HOME, label: 'Feed', icon: icons.home },
    { href: ROUTES.MESSAGES, label: 'Messages', icon: icons.messages },
    { href: ROUTES.NOTIFICATIONS, label: 'Alerts', icon: icons.notifications },
    { href: ROUTES.PROFILE(username), label: 'Profile', icon: icons.profile },
  ];

  return (
    <nav className={styles.bottomNav} aria-label="Main navigation">
      {items.map((item) => {
        const isActive =
          item.href === '/'
            ? pathname === '/'
            : pathname.startsWith(item.href);

        return (
          <Link
            key={item.href}
            href={item.href}
            className={`${styles.navItem} ${isActive ? styles.navItemActive : ''}`}
            aria-current={isActive ? 'page' : undefined}
          >
            {item.icon}
            <span>{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
