'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import LandingAuth from '@/components/auth/LandingAuth';
import FeedTabs from '@/components/feed/FeedTabs';
import PostComposer from '@/components/feed/PostComposer';
import PostCard from '@/components/feed/PostCard';
import { PostCardSkeleton } from '@/components/ui/Skeleton';
import Button from '@/components/ui/Button';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import type { Post } from '@/lib/types';

export default function HomePage() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const [activeTab, setActiveTab] = useState<'following' | 'foryou'>('foryou');
  const [posts, setPosts] = useState<Post[]>([]);
  const [page, setPage] = useState(1);
  const [feedLoading, setFeedLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);

  // Fetch feed posts when activeTab changes (only if authenticated)
  useEffect(() => {
    if (!isAuthenticated) return;
    let isSubscribed = true;

    api
      .get<Post[]>(`${API.FEED}?tab=${activeTab}&page=1`)
      .then((res) => {
        if (!isSubscribed) return;
        if (res.success && res.data) {
          setPosts(res.data);
          setHasMore(res.data.length > 0);
        }
        setPage(1);
        setFeedLoading(false);
      })
      .catch((err: unknown) => {
        if (!isSubscribed) return;
        console.error('Failed to fetch feed:', err);
        setFeedLoading(false);
      });

    return () => {
      isSubscribed = false;
    };
  }, [activeTab, isAuthenticated]);

  // Show nothing while checking initial auth status
  if (isLoading) {
    return null;
  }

  // If not logged in, render the Landing Auth screen directly on /
  if (!isAuthenticated) {
    return <LandingAuth />;
  }

  function handleLoadMore() {
    const nextPage = page + 1;
    setLoadingMore(true);
    api
      .get<Post[]>(`${API.FEED}?tab=${activeTab}&page=${nextPage}`)
      .then((res) => {
        if (res.success && res.data) {
          setPosts((prev) => [...prev, ...res.data]);
          setHasMore(res.data.length > 0);
          setPage(nextPage);
        }
      })
      .catch((err: unknown) => {
        console.error('Failed to load more posts:', err);
      })
      .finally(() => {
        setLoadingMore(false);
      });
  }

  function handlePostCreated(newPost: Post) {
    const postWithUser = {
      ...newPost,
      user: newPost.user?.username ? newPost.user : user!,
    };
    setPosts((prev) => [postWithUser, ...prev]);
  }

  function handlePostUpdate(updatedPost: Post) {
    setPosts((prev) =>
      prev.map((p) => (p.id === updatedPost.id ? updatedPost : p)),
    );
  }

  return (
    <main>
      <FeedTabs activeTab={activeTab} onTabChange={setActiveTab} />
      <PostComposer onPostCreated={handlePostCreated} />

      {feedLoading ? (
        <>
          <PostCardSkeleton />
          <PostCardSkeleton />
          <PostCardSkeleton />
        </>
      ) : posts.length === 0 ? (
        <div
          style={{
            padding: 'var(--space-6) var(--space-4)',
            textAlign: 'center',
            color: 'var(--text-muted)',
          }}
        >
          No posts yet — follow people to see their posts here, or switch to
          the For You tab.
        </div>
      ) : (
        <div>
          {posts.map((post, idx) => (
            <PostCard
              key={post.id ? `post-${post.id}` : `post-idx-${idx}`}
              post={post}
              onUpdate={handlePostUpdate}
            />
          ))}

          {hasMore && posts.length > 0 && (
            <div
              style={{
                padding: 'var(--space-4)',
                display: 'flex',
                justifyContent: 'center',
              }}
            >
              <Button
                variant="secondary"
                loading={loadingMore}
                onClick={handleLoadMore}
              >
                Load more
              </Button>
            </div>
          )}
        </div>
      )}
    </main>
  );
}
