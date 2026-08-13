'use client';

import { useState, useCallback } from 'react';
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
import { invalidateCache } from '@/lib/cache';
import { useCachedFetch } from '@/lib/useCachedFetch';
import type { Post } from '@/lib/types';
import { compressImage } from '@/lib/compression';

export default function HomePage() {
  const { isAuthenticated, isLoading } = useAuth();
  const { showToast } = useToast();
  const [activeTab, setActiveTab] = useState<'following' | 'foryou'>('foryou');
  const [page, setPage] = useState(1);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);

  // Use cached fetch for feed data - shows cached data immediately on repeat visits
  const { data: posts, loading: feedLoading } = useCachedFetch<Post[]>(
    `feed-${activeTab}-page1`,
    async () => {
      const res = await api.get<Post[]>(`${API.FEED}?tab=${activeTab}&page=1`);
      // Backend returns data directly as the posts array
      if (res.success && res.data) {
        return res.data;
      }
      return [];
    },
    {
      dependencies: [activeTab, isAuthenticated],
      onSuccess: (data) => {
        setHasMore(data.length > 0);
        setPage(1);
      },
    },
  );

  const safePosts = posts ?? [];

  /**
   * Called by PostComposer when the user clicks "Post".
   * Must be defined before early returns to satisfy rules-of-hooks.
   *
   * Flow:
   * 1. Upload files to R2 in background.
   * 2. Create the post record on the server.
   * 3. Invalidate cache so feed refreshes with new post.
   */
  const handlePostSubmit = useCallback(
    async (payload: {
      content: string;
      files: { file: File; type: 'photo' | 'video' }[];
      pendingKey: string;
    }) => {
      const { content, files } = payload;

      try {
        const mediaUrls: string[] = [];
        const mediaTypes: string[] = [];

        if (files.length > 0) {
          for (const { file, type } of files) {
            // Step 1: Compress image client-side if needed
            let uploadFile = file;
            if (type === 'photo' && file.size > 2 * 1024 * 1024) {
              try {
                uploadFile = await compressImage(file, {
                  maxSizeMB: 2,
                  maxWidthOrHeight: 2048,
                });
              } catch (e) {
                console.warn('[DEBUG] Compression failed, using original:', e);
              }
            }

            // Step 2: Request presigned URL from backend
            const requestRes = await api.post<{ url: string; key: string }>(
              API.MEDIA_UPLOAD_REQUEST,
              {
                filename: uploadFile.name,
                content_type: uploadFile.type,
                size: uploadFile.size,
              },
            );
            if (!requestRes.success || !requestRes.data?.url) {
              throw new Error(requestRes.error?.message || 'Failed to get presigned URL');
            }

            const { url: presignedUrl, key } = requestRes.data;

            // Step 3: Upload directly to R2 via presigned URL
            const putRes = await fetch(presignedUrl, {
              method: 'PUT',
              headers: {
                'Content-Type': uploadFile.type,
              },
              body: uploadFile,
            });

            if (!putRes.ok) {
              const errText = await putRes.text();
              throw new Error(`R2 upload failed: ${putRes.status} - ${errText.slice(0, 100)}`);
            }

            // Step 4: Confirm upload with backend
            const completeRes = await api.post<{ url: string }>(API.MEDIA_UPLOAD_COMPLETE, {
              key,
              size: uploadFile.size,
            });
            if (!completeRes.success || !completeRes.data?.url) {
              throw new Error(completeRes.error?.message || 'Upload confirmation failed');
            }

            mediaUrls.push(completeRes.data.url);
            mediaTypes.push(type);
          }
        }

        const res = await api.post<Post>(API.POSTS, {
          content,
          media_urls: mediaUrls,
          media_types: mediaTypes,
        });

        if (!res.success || !res.data) {
          throw new Error(res.error?.message || 'Failed to create post');
        }

        // Invalidate feed cache so it refreshes with the new post
        invalidateCache(`feed-${activeTab}-page1`);
      } catch (err) {
        console.error('Post creation failed:', err);
        // Show user-facing error message
        const errorMsg = err instanceof Error ? err.message : 'Failed to post. Please try again.';
        showToast(errorMsg, 'error');
      }
    },
    [activeTab, showToast],
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
      ) : safePosts.length === 0 ? (
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
          {safePosts.map((post, idx) => (
            <PostCard
              key={post._pendingKey ?? (post.id ? `post-${post.id}` : `post-idx-${idx}`)}
              post={post}
            />
          ))}

          {hasMore && safePosts.filter((p) => !p._pending).length > 0 && (
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
