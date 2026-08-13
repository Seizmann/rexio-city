/*
 * Search results page
 *
 * Implements full search results with tabs for Users / Posts / Hashtags.
 * DESIGN.md compliance: CSS custom properties only, mobile-first.
 */

'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import styles from './page.module.css';
import type { SearchResponse } from '@/lib/types';

type TabType = 'all' | 'user' | 'post' | 'hashtag';

const TABS: { key: TabType; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'user', label: 'Users' },
  { key: 'post', label: 'Posts' },
  { key: 'hashtag', label: 'Hashtags' },
];

export default function SearchPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const query = searchParams.get('q') || '';
  const typeParam = searchParams.get('type') || '';

  const [activeTab, setActiveTab] = useState<TabType>(
    typeParam === 'user' || typeParam === 'post' || typeParam === 'hashtag'
      ? typeParam
      : 'all'
  );
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  // Fetch search results
  const fetchResults = useCallback(
    async (searchQuery: string, searchType: TabType, pageNum: number) => {
      if (searchQuery.length < 2) {
        setResults(null);
        return;
      }

      setIsLoading(true);
      setError(null);

      try {
        const params = new URLSearchParams({
          q: searchQuery,
          page: String(pageNum),
          per_page: '10',
        });

        if (searchType !== 'all') {
          params.set('type', searchType);
        }

        const res = await api.get<SearchResponse>(`${API.SEARCH}?${params}`);

        if (res.success && res.data) {
          if (pageNum === 1) {
            setResults(res.data);
          } else {
            setResults((prev) => ({
              ...res.data,
              users: prev?.users ? [...prev.users, ...(res.data.users || [])] : res.data.users,
              posts: prev?.posts ? [...prev.posts, ...(res.data.posts || [])] : res.data.posts,
              hashtags: prev?.hashtags
                ? [...prev.hashtags, ...(res.data.hashtags || [])]
                : res.data.hashtags,
            }));
          }
          setHasMore((res.data?.users?.length || res.data?.posts?.length || res.data?.hashtags?.length || 0) >= 10);
        } else {
          setError(res.error?.message || 'Search failed');
        }
      } catch {
        setError('Network error. Please try again.');
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  // Use a ref to track the last fetched combination to avoid redundant fetches
  const lastFetchedRef = useRef<{ query: string; tab: TabType } | null>(null);

  // Fetch search results when query or tab changes
  useEffect(() => {
    const lastFetched = lastFetchedRef.current;
    if (lastFetched && lastFetched.query === query && lastFetched.tab === activeTab) {
      return; // Already fetched this combination
    }

    lastFetchedRef.current = { query, tab: activeTab };
    setPage(1);
    void fetchResults(query, activeTab, 1);
  }, [query, activeTab, fetchResults]);

  // Update URL when tab changes
  useEffect(() => {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (activeTab !== 'all') params.set('type', activeTab);
    router.replace(`/search?${params.toString()}`, { scroll: false });
  }, [activeTab, query, router]);

  const handleLoadMore = () => {
    const nextPage = page + 1;
    setPage(nextPage);
    void fetchResults(query, activeTab, nextPage);
  };

  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab);
  };

  if (!query) {
    return (
      <main className={styles.main}>
        <div className={styles.emptyState}>
          <h1 className={styles.srOnly}>Search</h1>
          <p>Enter a search query to find users, posts, or hashtags.</p>
        </div>
      </main>
    );
  }

  return (
    <main className={styles.main}>
      <div className={styles.header}>
        <h1 className={styles.title}>Search</h1>
        <p className={styles.subtitle}>
          {isLoading ? 'Searching...' : `Results for "${query}"`}
        </p>
      </div>

      {/* Tabs */}
      <div className={styles.tabs} role="tablist" aria-label="Search result types">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            className={`${styles.tab} ${activeTab === tab.key ? styles.tabActive : ''}`}
            onClick={() => handleTabChange(tab.key)}
            role="tab"
            aria-selected={activeTab === tab.key}
            aria-controls={`panel-${tab.key}`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Error */}
      {error && (
        <div className={styles.error} role="alert">
          {error}
        </div>
      )}

      {/* Results */}
      {activeTab === 'all' && (
        <div role="tabpanel" id="panel-all">
          {/* Users */}
          {results?.has_users && results.users && results.users.length > 0 && (
            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>Users</h2>
              <div className={styles.userList}>
                {results.users.map((user) => (
                  <Link
                    key={user.id}
                    href={`/${user.username}`}
                    className={styles.userCard}
                  >
                    {user.avatar_url ? (
                      <img
                        src={user.avatar_url}
                        alt=""
                        className={styles.avatar}
                      />
                    ) : (
                      <div className={styles.avatarFallback}>
                        {user.username.slice(0, 2).toUpperCase()}
                      </div>
                    )}
                    <div className={styles.userInfo}>
                      <span className={styles.username}>@{user.username}</span>
                      {user.display_name && (
                        <span className={styles.displayName}>{user.display_name}</span>
                      )}
                    </div>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Posts */}
          {results?.has_posts && results.posts && results.posts.length > 0 && (
            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>Posts</h2>
              <div className={styles.postList}>
                {results.posts.map((post) => (
                  <Link
                    key={post.id}
                    href={`/post/${post.public_id}`}
                    className={styles.postCard}
                  >
                    <div className={styles.postHeader}>
                      {post.user.avatar_url ? (
                        <img
                          src={post.user.avatar_url}
                          alt=""
                          className={styles.avatarSmall}
                        />
                      ) : (
                        <div className={styles.avatarFallbackSmall}>
                          {post.user.username.slice(0, 2).toUpperCase()}
                        </div>
                      )}
                      <span className={styles.postUser}>@{post.user.username}</span>
                    </div>
                    <p className={styles.postContent}>{post.content}</p>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Hashtags */}
          {results?.has_hashtags && results.hashtags && results.hashtags.length > 0 && (
            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>Hashtags</h2>
              <div className={styles.hashtagList}>
                {results.hashtags.map((htag) => (
                  <Link
                    key={htag.hashtag}
                    href={`/search?q=${encodeURIComponent(htag.hashtag)}&type=hashtag`}
                    className={styles.hashtagCard}
                  >
                    <span className={styles.hashtag}>{htag.hashtag}</span>
                    <span className={styles.hashtagCount}>{htag.count} posts</span>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Empty state */}
          {!results?.has_users &&
            !results?.has_posts &&
            !results?.has_hashtags &&
            !isLoading && (
              <div className={styles.emptyState}>
                <p>No results for &ldquo;{query}&rdquo;</p>
              </div>
            )}
        </div>
      )}

      {activeTab === 'user' && results?.users && results.users.length > 0 && (
        <div role="tabpanel" id="panel-user">
          <section className={styles.section}>
            <div className={styles.userList}>
              {results.users.map((user) => (
                <Link
                  key={user.id}
                  href={`/${user.username}`}
                  className={styles.userCard}
                >
                  {user.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt=""
                      className={styles.avatar}
                    />
                  ) : (
                    <div className={styles.avatarFallback}>
                      {user.username.slice(0, 2).toUpperCase()}
                    </div>
                  )}
                  <div className={styles.userInfo}>
                    <span className={styles.username}>@{user.username}</span>
                    {user.display_name && (
                      <span className={styles.displayName}>{user.display_name}</span>
                    )}
                  </div>
                </Link>
              ))}
            </div>
          </section>
        </div>
      )}

      {activeTab === 'post' && results?.posts && results.posts.length > 0 && (
        <div role="tabpanel" id="panel-post">
          <section className={styles.section}>
            <div className={styles.postList}>
              {results.posts.map((post) => (
                <Link
                  key={post.id}
                  href={`/post/${post.public_id}`}
                  className={styles.postCard}
                >
                  <div className={styles.postHeader}>
                    {post.user.avatar_url ? (
                      <img
                        src={post.user.avatar_url}
                        alt=""
                        className={styles.avatarSmall}
                      />
                    ) : (
                      <div className={styles.avatarFallbackSmall}>
                        {post.user.username.slice(0, 2).toUpperCase()}
                      </div>
                    )}
                    <span className={styles.postUser}>@{post.user.username}</span>
                  </div>
                  <p className={styles.postContent}>{post.content}</p>
                </Link>
              ))}
            </div>
          </section>
        </div>
      )}

      {activeTab === 'hashtag' && results?.hashtags && results.hashtags.length > 0 && (
        <div role="tabpanel" id="panel-hashtag">
          <section className={styles.section}>
            <div className={styles.hashtagList}>
              {results.hashtags.map((htag) => (
                <Link
                  key={htag.hashtag}
                  href={`/search?q=${encodeURIComponent(htag.hashtag)}&type=hashtag`}
                  className={styles.hashtagCard}
                >
                  <span className={styles.hashtag}>{htag.hashtag}</span>
                  <span className={styles.hashtagCount}>{htag.count} posts</span>
                </Link>
              ))}
            </div>
          </section>
        </div>
      )}

      {/* Load more */}
      {!isLoading && hasMore && results && (
        <div className={styles.loadMore}>
          <button className={styles.loadMoreButton} onClick={handleLoadMore}>
            Load more
          </button>
        </div>
      )}

      {/* Loading state */}
      {isLoading && (
        <div className={styles.loading} role="status">
          <span className={styles.srOnly}>Loading results...</span>
        </div>
      )}
    </main>
  );
}
