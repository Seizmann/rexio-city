'use client';

import { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { useAuth } from '@/context/AuthContext';
import { ROUTES } from '@/lib/constants';
import SearchInput from '@/components/search/SearchInput';
import styles from './TopBar.module.css';

/**
 * Top navigation bar — logo (left), user avatar with dropdown (right).
 * DESIGN.md §4: "search input, a few status/utility icons, user avatar."
 * MVP: search is deferred; just logo + avatar dropdown.
 */
export default function TopBar() {
  const { user, logout } = useAuth();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    }
    if (dropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [dropdownOpen]);

  const initials = user?.display_name
    ? user.display_name.slice(0, 2)
    : user?.username?.slice(0, 2) || '?';

  return (
    <header className={styles.topbar}>
      <Link href={ROUTES.HOME} className={styles.logo}>
        <img src="/icon.webp" alt="RexiO City Icon" className={styles.logoIcon} />
        <span>RexiO City</span>
      </Link>

      <div className={styles.actions}>
        <SearchInput className={styles.searchInput} />
        <div ref={dropdownRef} style={{ position: 'relative' }}>
          <button
            className={styles.avatarBtn}
            onClick={() => setDropdownOpen(!dropdownOpen)}
            aria-label="Account menu"
            aria-expanded={dropdownOpen}
          >
            {user?.avatar_url ? (
              <img
                src={user.avatar_url}
                alt={user.display_name || user.username}
                className={styles.avatarImg}
              />
            ) : (
              <span className={styles.avatarFallback}>{initials}</span>
            )}
          </button>

          {dropdownOpen && (
            <div className={styles.dropdown} role="menu">
              <Link
                href={ROUTES.PROFILE(user?.username || '')}
                className={styles.dropdownItem}
                role="menuitem"
                onClick={() => setDropdownOpen(false)}
              >
                Profile
              </Link>
              <div className={styles.dropdownDivider} />
              <button
                className={`${styles.dropdownItem} ${styles.dropdownDanger}`}
                role="menuitem"
                onClick={() => {
                  setDropdownOpen(false);
                  void logout();
                }}
              >
                Log out
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
