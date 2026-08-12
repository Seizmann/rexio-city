'use client';

import React, { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import styles from './page.module.css';
import { api } from '@/lib/api';
import { API, ROUTES } from '@/lib/constants';
import type { Post, Comment, User } from '@/lib/types';
import PostCard from '@/components/feed/PostCard';
import { PostCardSkeleton } from '@/components/ui/Skeleton';
import Button from '@/components/ui/Button';
import Link from 'next/link';
import { relativeTime } from '@/lib/time';
import { useAuth } from '@/context/AuthContext';

interface GetPostOutput {
  post: Post;
  likes: number;
  comments: number;
  reposts: number;
  is_liked: boolean;
  is_reposted: boolean;
  is_bookmarked: boolean;
}

export default function PostDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { user: authUser } = useAuth();
  const postId = params.id as string;

  const [postData, setPostData] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const [newComment, setNewComment] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!postId) return;
    let isSubscribed = true;

    Promise.all([
      api.get<GetPostOutput | Post>(API.POST(postId)),
      api.get<Comment[]>(API.POST_COMMENTS(postId)),
    ])
      .then(([postRes, commentsRes]) => {
        if (!isSubscribed) return;

        if (postRes.success && postRes.data) {
          const raw = postRes.data as GetPostOutput;
          if (raw.post) {
            const formattedPost: Post = {
              ...raw.post,
              likes: raw.likes,
              like_count: raw.likes,
              comments: raw.comments,
              comment_count: raw.comments,
              reposts: raw.reposts,
              repost_count: raw.reposts,
              is_liked: raw.is_liked,
              is_reposted: raw.is_reposted,
              is_bookmarked: raw.is_bookmarked,
            };
            setPostData(formattedPost);
          } else {
            setPostData(postRes.data as Post);
          }
        } else {
          setError(true);
        }

        if (commentsRes.success && commentsRes.data) {
          setComments(commentsRes.data);
        }
        setLoading(false);
      })
      .catch((err: unknown) => {
        if (!isSubscribed) return;
        console.error('Failed to load post details:', err);
        setError(true);
        setLoading(false);
      });

    return () => {
      isSubscribed = false;
    };
  }, [postId]);

  const handlePostUpdate = (updatedPost: Post) => {
    setPostData(updatedPost);
  };

  const handleCommentSubmit = async () => {
    if (!newComment.trim() || isSubmitting || !postId) return;

    setIsSubmitting(true);
    try {
      const res = await api.post<Comment>(API.POST_COMMENTS(postId), {
        content: newComment,
      });

      if (res.success && res.data) {
        const created = res.data;
        if (!created.user && authUser) {
          created.user = authUser;
        }
        setComments((prev) => [...prev, created]);
        setNewComment('');

        // Increment post comment count
        if (postData) {
          const currentCount = postData.comment_count ?? postData.comments ?? 0;
          setPostData({
            ...postData,
            comments: currentCount + 1,
            comment_count: currentCount + 1,
          });
        }
      }
    } catch (err: unknown) {
      console.error('Failed to submit comment:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <button
          className={styles.backBtn}
          onClick={() => router.back()}
          aria-label="Go back"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="19" y1="12" x2="5" y2="12" />
            <polyline points="12 19 5 12 12 5" />
          </svg>
        </button>
        <h1 className={styles.title}>Post</h1>
      </header>

      {loading ? (
        <PostCardSkeleton />
      ) : error || !postData ? (
        <div className={styles.notFound}>
          <h2>Post Not Found</h2>
          <p>This post may have been deleted or does not exist.</p>
        </div>
      ) : (
        <>
          <PostCard post={postData} onUpdate={handlePostUpdate} />

          <section className={styles.commentsSection}>
            <h2 className={styles.commentsTitle}>Replies</h2>

            {authUser && (
              <div className={styles.composer}>
                <textarea
                  className={styles.textarea}
                  placeholder="Post your reply..."
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  disabled={isSubmitting}
                />
                <div className={styles.composerAction}>
                  <Button
                    variant="primary"
                    size="sm"
                    loading={isSubmitting}
                    disabled={!newComment.trim()}
                    onClick={() => {
                      void handleCommentSubmit();
                    }}
                  >
                    Reply
                  </Button>
                </div>
              </div>
            )}

            {comments.length === 0 ? (
              <div className={styles.empty}>
                No replies yet. Be the first to join the conversation!
              </div>
            ) : (
              <div className={styles.commentList}>
                {comments.map((comment) => {
                  const author: User =
                    comment.user && comment.user.username
                      ? comment.user
                      : authUser && authUser.id === comment.user_id
                      ? authUser
                      : {
                          id: comment.user_id,
                          username: comment.user?.username || 'user',
                          display_name:
                            comment.user?.display_name ||
                            comment.user?.username ||
                            'User',
                        };

                  return (
                    <div key={comment.id} className={styles.comment}>
                      <Link href={ROUTES.PROFILE(author.username)}>
                        {author.avatar_url ? (
                          <img
                            src={author.avatar_url}
                            alt={author.display_name || author.username}
                            className={styles.avatar}
                          />
                        ) : (
                          <div className={styles.avatarFallback}>
                            {author.display_name?.[0] ||
                              author.username?.[0] ||
                              '?'}
                          </div>
                        )}
                      </Link>

                      <div className={styles.commentBody}>
                        <div className={styles.authorRow}>
                          <Link href={ROUTES.PROFILE(author.username)}>
                            <span className={styles.displayName}>
                              {author.display_name || author.username}
                            </span>
                          </Link>
                          <span className={styles.username}>
                            @{author.username}
                          </span>
                          <span className={styles.timestamp}>
                            {relativeTime(comment.created_at)}
                          </span>
                        </div>
                        <div className={styles.text}>{comment.content}</div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </section>
        </>
      )}
    </div>
  );
}
