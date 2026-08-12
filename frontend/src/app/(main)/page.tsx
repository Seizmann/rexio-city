'use client';

import { useState, useEffect, useCallback } from 'react';
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
    setFeedLoading(true);

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

  if (isLoading) return null;
  if (!isAuthenticated) return <LandingAuth />;

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

  function handlePostUpdate(updatedPost: Post) {
    setPosts((prev) =>
      prev.map((p) => (p.id === updatedPost.id ? updatedPost : p)),
    );
  }

  /**
   * Called by PostComposer when the user clicks "Post".
   *
   * Flow:
   * 1. Instantly add a pending placeholder at the top of the feed (with local
   *    preview blobs and "Uploading…" status). The user sees content immediately.
   * 2. In the background: upload each file to R2, then create the post.
   * 3. Update the placeholder's status text as we progress through stages.
   * 4. Replace the placeholder with the real confirmed post from the server.
   */
  const handlePostSubmit = useCallback(
    async (payload: {
      content: string;
      files: { file: File; type: 'photo' | 'video' }[];
      pendingKey: string;
    }) => {
      const { content, files, pendingKey } = payload;

      // Build local preview blobs (no upload yet)
      const localPreviews = files.map(({ file, type }) => ({
        previewUrl: URL.createObjectURL(file),
        type,
      }));

      // Ensure the user object is fully resolved so the card never shows blank
      const resolvedUser = user
        ? {
            id: user.id,
            username: user.username,
            display_name: user.display_name,
            avatar_url: user.avatar_url,
          }
        : undefined;

      // 1. Optimistic placeholder at the top of the feed
      const pendingPost: Post = {
        id: 0, // will be replaced
        user_id: user?.id ?? 0,
        user: resolvedUser,
        content,
        media: [],
        created_at: new Date().toISOString(),
        _pending: true,
        _uploadStatus: files.length > 0 ? 'uploading' : 'finishing',
        _localPreviews: localPreviews,
        _pendingKey: pendingKey,
      };

      setPosts((prev) => [pendingPost, ...prev]);

      /**
       * Update the status label on the pending card in the feed.
       * We match by _pendingKey to avoid touching other cards.
       */
      const updateStatus = (status: Post['_uploadStatus']) => {
        setPosts((prev) =>
          prev.map((p) =>
            p._pendingKey === pendingKey ? { ...p, _uploadStatus: status } : p,
          ),
        );
      };

      try {
        // 2. Upload files if any (shows "Uploading…")
        const mediaUrls: string[] = [];
        const mediaTypes: string[] = [];

        if (files.length > 0) {
          for (const { file, type } of files) {
            const formData = new FormData();
            formData.append('file', file);
            const res = await api.upload<{ url: string; type: string }>(
              API.MEDIA_UPLOAD,
              formData,
            );
            if (!res.success || !res.data?.url) {
              throw new Error(res.error?.message || 'Media upload failed');
            }
            mediaUrls.push(res.data.url);
            mediaTypes.push(res.data.type || type);
          }
          // 3. "Updating…" — files uploaded, now creating the post record
          updateStatus('updating');
        }

        // Small perceived-progress pause so "Finishing…" is visible briefly
        updateStatus('finishing');

        // 4. Create the post on the server
        const res = await api.post<Post>(API.POSTS, {
          content,
          media_urls: mediaUrls,
          media_types: mediaTypes,
        });

        if (!res.success || !res.data) {
          throw new Error(res.error?.message || 'Failed to create post');
        }

        // 5. Replace the pending placeholder with the real confirmed post.
        //    Ensure the user object is populated (backend may return minimal user).
        const confirmedPost: Post = {
          ...res.data,
          user: res.data.user?.username ? res.data.user : resolvedUser,
        };

        // Revoke local preview blob URLs now that we have real URLs
        localPreviews.forEach((p) => URL.revokeObjectURL(p.previewUrl));

        setPosts((prev) =>
          prev.map((p) =>
            p._pendingKey === pendingKey ? confirmedPost : p,
          ),
        );
      } catch (err) {
        console.error('Post creation failed:', err);
        // Revoke blobs on error too
        localPreviews.forEach((p) => URL.revokeObjectURL(p.previewUrl));
        updateStatus('error');
        // Remove the failed placeholder after 3 seconds
        setTimeout(() => {
          setPosts((prev) => prev.filter((p) => p._pendingKey !== pendingKey));
        }, 3000);
      }
    },
    [user],
  );

  return (
    <main>
      <FeedTabs activeTab={activeTab} onTabChange={setActiveTab} />
      <PostComposer onPostSubmit={handlePostSubmit} />

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
              key={post._pendingKey ?? (post.id ? `post-${post.id}` : `post-idx-${idx}`)}
              post={post}
              onUpdate={handlePostUpdate}
            />
          ))}

          {hasMore && posts.filter((p) => !p._pending).length > 0 && (
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
