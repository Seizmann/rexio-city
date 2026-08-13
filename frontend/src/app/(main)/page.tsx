'use client';

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import LandingAuth from '@/components/auth/LandingAuth';
import FeedTabs from '@/components/feed/FeedTabs';
import PostComposer from '@/components/feed/PostComposer';
import PostCard from '@/components/feed/PostCard';
import { PostCardSkeleton } from '@/components/ui/Skeleton';
import Button from '@/components/ui/Button';
import { useToast } from '@/components/ui/Toast';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import type { Post } from '@/lib/types';
import DebugPanel, { debugLog } from '@/components/debug/DebugPanel';

export default function HomePage() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const { showToast } = useToast();
  const [activeTab, setActiveTab] = useState<'following' | 'foryou'>('foryou');
  const [posts, setPosts] = useState<Post[]>([]);
  const [page, setPage] = useState(1);
  const [feedLoading, setFeedLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);

  // Fetch feed posts when activeTab changes (only if authenticated).
  // setFeedLoading is called inside the async callback, not synchronously
  // in the effect body, to satisfy react-hooks/set-state-in-effect.
  useEffect(() => {
    if (!isAuthenticated) return;
    let isSubscribed = true;

    // Kick off loading state via a microtask so it's not synchronous in the effect
    void Promise.resolve().then(() => {
      if (isSubscribed) setFeedLoading(true);
    });

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

  /**
   * Called by PostComposer when the user clicks "Post".
   * Must be defined before early returns to satisfy rules-of-hooks.
   *
   * Flow:
   * 1. Instantly add a pending placeholder at the top of the feed.
   * 2. Upload files to R2 in background, updating status label.
   * 3. Create the post record on the server.
   * 4. Replace the placeholder with the confirmed post.
   */
  const handlePostSubmit = useCallback(
    async (payload: {
      content: string;
      files: { file: File; type: 'photo' | 'video' }[];
      pendingKey: string;
    }) => {
      const { content, files, pendingKey } = payload;

      debugLog('upload', `Starting post: ${files.length} file(s), content: ${content.slice(0, 30)}...`);
      debugLog('auth', `User: ${user?.username}, Token: ${!!api.getAccessToken()}`);

      const localPreviews = files.map(({ file, type }) => ({
        previewUrl: URL.createObjectURL(file),
        type,
      }));

      const resolvedUser = user
        ? {
            id: user.id,
            username: user.username,
            display_name: user.display_name,
            avatar_url: user.avatar_url,
          }
        : undefined;

      const pendingPost: Post = {
        id: 0,
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

      const updateStatus = (status: Post['_uploadStatus']) => {
        setPosts((prev) =>
          prev.map((p) =>
            p._pendingKey === pendingKey ? { ...p, _uploadStatus: status } : p,
          ),
        );
      };

      try {
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
          updateStatus('updating');
        }

        updateStatus('finishing');

        const res = await api.post<Post & { post?: Post }>(API.POSTS, {
          content,
          media_urls: mediaUrls,
          media_types: mediaTypes,
        });

        if (!res.success || !res.data) {
          throw new Error(res.error?.message || 'Failed to create post');
        }

        // Safely unwrap raw post object whether backend returns { post: Post } or Post directly
        const rawPost: Post = (res.data.post && typeof res.data.post === 'object') ? res.data.post : res.data;

        const confirmedPost: Post = {
          ...rawPost,
          user: rawPost.user?.username ? rawPost.user : resolvedUser,
        };

        localPreviews.forEach((p) => URL.revokeObjectURL(p.previewUrl));

        setPosts((prev) =>
          prev.map((p) =>
            p._pendingKey === pendingKey ? confirmedPost : p,
          ),
        );
      } catch (err) {
        console.error('Post creation failed:', err);
        debugLog('error', `Post failed: ${err instanceof Error ? err.message : String(err)}`);
        localPreviews.forEach((p) => URL.revokeObjectURL(p.previewUrl));
        updateStatus('error');
        // Show user-facing error message
        const errorMsg = err instanceof Error ? err.message : 'Failed to post. Please try again.';
        showToast(errorMsg, 'error');
        setTimeout(() => {
          setPosts((prev) => prev.filter((p) => p._pendingKey !== pendingKey));
        }, 3000);
      }
    },
    [user],
  );

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

  return (
    <main>
      <FeedTabs activeTab={activeTab} onTabChange={setActiveTab} />
      <PostComposer onPostSubmit={(p) => { void handlePostSubmit(p); }} />

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
