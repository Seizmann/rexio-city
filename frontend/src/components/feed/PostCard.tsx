'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import styles from './PostCard.module.css';
import { api } from '@/lib/api';
import { API, ROUTES } from '@/lib/constants';
import type { Post, User } from '@/lib/types';
import { relativeTime } from '@/lib/time';
import CommentSheet from './CommentSheet';

interface PostCardProps {
  post: Post;
  onUpdate?: (post: Post) => void;
}

export default function PostCard({ post, onUpdate }: PostCardProps) {
  const router = useRouter();
  const [localPost, setLocalPost] = useState<Post>(post);
  const [isCommentSheetOpen, setIsCommentSheetOpen] = useState(false);

  // Author fallback if user relationship is undefined
  const author: User = localPost.user || {
    id: localPost.user_id,
    username: 'user',
    display_name: 'Anonymous',
  };

  const likeCount = localPost.like_count ?? localPost.likes ?? 0;
  const commentCount = localPost.comment_count ?? localPost.comments ?? 0;
  const repostCount = localPost.repost_count ?? localPost.reposts ?? 0;
  const bookmarkCount = localPost.bookmark_count ?? 0;

  const handleCardClick = (e: React.MouseEvent) => {
    const target = e.target as HTMLElement;
    if (target.closest('button') || target.closest('a') || target.closest('textarea')) {
      return;
    }
    router.push(ROUTES.POST(post.id));
  };

  const handleAction = async (
    actionType: 'like' | 'repost' | 'bookmark',
    apiEndpoint: string | ((id: number) => string)
  ) => {
    const isCurrentlyActive = !!localPost[`is_${actionType}d` as keyof Post];
    const updatedIsActive = !isCurrentlyActive;

    const countDelta = updatedIsActive ? 1 : -1;
    const nextLikes = likeCount + (actionType === 'like' ? countDelta : 0);
    const nextComments = commentCount;
    const nextReposts = repostCount + (actionType === 'repost' ? countDelta : 0);
    const nextBookmarks = bookmarkCount + (actionType === 'bookmark' ? countDelta : 0);

    // Optimistic update
    const updatedPost: Post = {
      ...localPost,
      [`is_${actionType}d`]: updatedIsActive,
      likes: Math.max(0, nextLikes),
      like_count: Math.max(0, nextLikes),
      comments: Math.max(0, nextComments),
      comment_count: Math.max(0, nextComments),
      reposts: Math.max(0, nextReposts),
      repost_count: Math.max(0, nextReposts),
      bookmark_count: Math.max(0, nextBookmarks),
    };
    setLocalPost(updatedPost);
    if (onUpdate) onUpdate(updatedPost);

    try {
      const endpoint = typeof apiEndpoint === 'function' ? apiEndpoint(post.id) : apiEndpoint;
      if (isCurrentlyActive) {
        await api.delete(endpoint);
      } else {
        await api.post(endpoint);
      }
    } catch (error: unknown) {
      console.error(`Failed to toggle ${actionType}:`, error);
      // Revert on error
      setLocalPost(localPost);
      if (onUpdate) onUpdate(localPost);
    }
  };

  const handleCommentAdded = () => {
    const nextCommentCount = commentCount + 1;
    const updatedPost: Post = {
      ...localPost,
      comments: nextCommentCount,
      comment_count: nextCommentCount,
    };
    setLocalPost(updatedPost);
    if (onUpdate) onUpdate(updatedPost);
  };

  return (
    <>
      <article className={styles.card} onClick={handleCardClick}>
        <Link href={ROUTES.PROFILE(author.username)}>
          {author.avatar_url ? (
            <img src={author.avatar_url} alt={author.display_name} className={styles.avatar} />
          ) : (
            <div className={styles.avatarFallback}>{author.display_name?.[0] || author.username?.[0] || '?'}</div>
          )}
        </Link>
        <div className={styles.content}>
          <div className={styles.authorRow}>
            <Link href={ROUTES.PROFILE(author.username)} className={styles.authorLink}>
              <span className={styles.displayName}>{author.display_name || author.username}</span>
              <span className={styles.username}>@{author.username}</span>
            </Link>
            <span className={styles.timestamp}>{relativeTime(localPost.created_at)}</span>
          </div>

          <div className={styles.body}>{localPost.content}</div>

          <div className={styles.actionRow}>
            <button
              className={`${styles.actionButton} ${localPost.is_liked ? styles.active : ''}`}
              onClick={() => { void handleAction('like', API.POST_LIKE(post.id)); }}
              aria-label={localPost.is_liked ? 'Unlike' : 'Like'}
            >
              <svg viewBox="0 0 24 24" fill={localPost.is_liked ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
              </svg>
              <span>{likeCount > 0 ? likeCount : 0}</span>
            </button>

            <button
              className={styles.actionButton}
              onClick={() => setIsCommentSheetOpen(true)}
              aria-label="Comment"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
              </svg>
              <span>{commentCount > 0 ? commentCount : 0}</span>
            </button>

            <button
              className={`${styles.actionButton} ${localPost.is_reposted ? styles.active : ''}`}
              onClick={() => { void handleAction('repost', API.POST_REPOST(post.id)); }}
              aria-label={localPost.is_reposted ? 'Undo Repost' : 'Repost'}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="17 1 21 5 17 9" />
                <path d="M3 11V9a4 4 0 0 1 4-4h14" />
                <polyline points="7 23 3 19 7 15" />
                <path d="M21 13v2a4 4 0 0 1-4 4H3" />
              </svg>
              <span>{repostCount > 0 ? repostCount : 0}</span>
            </button>

            <button
              className={`${styles.actionButton} ${localPost.is_bookmarked ? styles.active : ''}`}
              onClick={() => { void handleAction('bookmark', API.POST_BOOKMARK(post.id)); }}
              aria-label={localPost.is_bookmarked ? 'Remove Bookmark' : 'Bookmark'}
            >
              <svg viewBox="0 0 24 24" fill={localPost.is_bookmarked ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
              </svg>
              <span>{bookmarkCount > 0 ? bookmarkCount : 0}</span>
            </button>
          </div>
        </div>
      </article>

      <CommentSheet
        postId={post.id}
        isOpen={isCommentSheetOpen}
        onClose={() => setIsCommentSheetOpen(false)}
        onCommentAdded={handleCommentAdded}
      />
    </>
  );
}
