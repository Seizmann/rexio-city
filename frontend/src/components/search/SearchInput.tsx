/*
 * RexiO City — Search UI Component
 *
 * Implements a debounced search input in the TopBar with a dropdown
 * showing combined results (users, posts, hashtags).
 *
 * DESIGN.md compliance:
 * - Uses CSS custom properties exclusively (no hardcoded colors/spacing)
 * - Accessible: <label> (visually hidden), aria-labels, focus-visible outlines
 */

'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import styles from './SearchInput.module.css';
import type { SearchResponse, SearchUserResult, SearchPostResult, SearchHashtagResult } from '@/lib/types';

interface SearchInputProps {
  className?: string;
}

export default function SearchInput({ className }: SearchInputProps) {
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Debounced search
  const debouncedSearch = useCallback(
    async (searchQuery: string) => {
      if (searchQuery.length < 2) {
        setResults(null);
        setIsOpen(false);
        return;
      }

      setIsLoading(true);
      setError(null);

      try {
        const res = await api.get<SearchResponse>(
          `${API.SEARCH}?q=${encodeURIComponent(searchQuery)}`
        );

        if (res.success && res.data) {
          setResults(res.data);
          setIsOpen(true);
        } else {
          setError(res.error?.message || 'Search failed');
          setIsOpen(false);
        }
      } catch {
        setError('Network error. Please try again.');
        setIsOpen(false);
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  // Cleanup query on unmount
  useEffect(() => {
    return () => {
      setQuery('');
      setResults(null);
    };
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    // Debounce the search
    const timeoutId = setTimeout(() => {
      void debouncedSearch(value);
    }, 300);
    return () => clearTimeout(timeoutId);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Escape') {
      setIsOpen(false);
      inputRef.current?.blur();
    } else if (e.key === 'Enter' && query.length >= 2) {
      // Navigate to full search results page
      router.push(`/search?q=${encodeURIComponent(query)}`);
      setIsOpen(false);
    }
  };

  const handleClear = () => {
    setQuery('');
    setResults(null);
    setIsOpen(false);
    inputRef.current?.focus();
  };

  const handleResultClick = () => {
    setIsOpen(false);
    setQuery('');
  };

  if (!isOpen && !query) {
    return (
      <div className={`${styles.container} ${className}`}>
        <label htmlFor="search-input" className={styles.srOnly}>
          Search RexiO City
        </label>
        <button
          className={styles.searchButton}
          onClick={() => {
            setIsOpen(true);
            inputRef.current?.focus();
          }}
          aria-label="Search"
        >
          <svg
            className={styles.searchIcon}
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.35-4.35" />
          </svg>
          <span className={styles.searchPlaceholder}>Search</span>
        </button>
      </div>
    );
  }

  return (
    <div ref={containerRef} className={`${styles.container} ${styles.open} ${className}`}>
      <label htmlFor="search-input" className={styles.srOnly}>
        Search RexiO City
      </label>
      <div className={styles.searchBar}>
        <svg
          className={styles.searchIcon}
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </svg>
        <input
          ref={inputRef}
          id="search-input"
          type="text"
          value={query}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          className={styles.searchInput}
          placeholder="Search users, posts, hashtags..."
          aria-autocomplete="list"
          aria-controls="search-results"
          aria-expanded={isOpen}
          aria-haspopup="listbox"
        />
        {query && (
          <button
            className={styles.clearButton}
            onClick={handleClear}
            aria-label="Clear search"
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        )}
      </div>

      {isOpen && (
        <div
          id="search-results"
          className={styles.results}
          role="listbox"
          aria-label="Search results"
        >
          {isLoading && (
            <div className={styles.loading} role="status">
              <span className={styles.srOnly}>Searching...</span>
            </div>
          )}

          {!isLoading && error && (
            <div className={styles.error} role="alert">
              {error}
            </div>
          )}

          {!isLoading && !error && results && (
            <>
              {/* Users */}
              {results.has_users && results.users && results.users.length > 0 && (
                <div className={styles.resultSection} role="group">
                  <div className={styles.resultSectionTitle}>Users</div>
                  <ul className={styles.resultList}>
                    {results.users.map((user: SearchUserResult) => (
                      <li key={user.id} role="option">
                        <Link
                          href={`/${user.username}`}
                          className={styles.resultItem}
                          onClick={handleResultClick}
                        >
                          {user.avatar_url ? (
                            <img
                              src={user.avatar_url}
                              alt=""
                              className={styles.resultAvatar}
                            />
                          ) : (
                            <div className={styles.resultAvatarFallback}>
                              {user.username.slice(0, 2).toUpperCase()}
                            </div>
                          )}
                          <div className={styles.resultInfo}>
                            <span className={styles.resultUsername}>@{user.username}</span>
                            {user.display_name && (
                              <span className={styles.resultDisplayName}>{user.display_name}</span>
                            )}
                          </div>
                        </Link>
                      </li>
                    ))}
                  </ul>
                  <Link
                    href={`/search?q=${encodeURIComponent(query)}&type=user`}
                    className={styles.seeAll}
                    onClick={handleResultClick}
                  >
                    See all users
                  </Link>
                </div>
              )}

              {/* Posts */}
              {results.has_posts && results.posts && results.posts.length > 0 && (
                <div className={styles.resultSection} role="group">
                  <div className={styles.resultSectionTitle}>Posts</div>
                  <ul className={styles.resultList}>
                    {results.posts.map((post: SearchPostResult) => (
                      <li key={post.id} role="option">
                        <Link
                          href={`/post/${post.public_id}`}
                          className={styles.resultItem}
                          onClick={handleResultClick}
                        >
                          <div className={styles.resultContent}>
                            <span className={styles.resultPostContent}>{post.content}</span>
                            <span className={styles.resultPostUser}>
                              @{post.user.username}
                            </span>
                          </div>
                        </Link>
                      </li>
                    ))}
                  </ul>
                  <Link
                    href={`/search?q=${encodeURIComponent(query)}&type=post`}
                    className={styles.seeAll}
                    onClick={handleResultClick}
                  >
                    See all posts
                  </Link>
                </div>
              )}

              {/* Hashtags */}
              {results.has_hashtags && results.hashtags && results.hashtags.length > 0 && (
                <div className={styles.resultSection} role="group">
                  <div className={styles.resultSectionTitle}>Hashtags</div>
                  <ul className={styles.resultList}>
                    {results.hashtags.map((htag: SearchHashtagResult) => (
                      <li key={htag.hashtag} role="option">
                        <Link
                          href={`/search?q=${encodeURIComponent(htag.hashtag)}&type=hashtag`}
                          className={styles.resultItem}
                          onClick={handleResultClick}
                        >
                          <span className={styles.hashtag}>{htag.hashtag}</span>
                          <span className={styles.hashtagCount}>{htag.count} posts</span>
                        </Link>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {/* Empty state */}
              {!results.has_users && !results.has_posts && !results.has_hashtags && (
                <div className={styles.empty} role="status">
                  No results for &ldquo;{query}&rdquo;
                </div>
              )}
            </>
          )}

          {!isLoading && !error && !results && (
            <div className={styles.loading} role="status">
              <span className={styles.srOnly}>Searching...</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
