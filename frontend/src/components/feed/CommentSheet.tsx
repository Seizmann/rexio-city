'use client';

import React, { useState, useEffect } from 'react';
import styles from './CommentSheet.module.css';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import type { Comment } from '@/lib/types';
import { relativeTime } from '@/lib/time';
import Button from '@/components/ui/Button';

interface CommentSheetProps {
  postId: number;
  isOpen: boolean;
  onClose: () => void;
  onCommentAdded?: () => void;
}

export default function CommentSheet({
  postId,
  isOpen,
  onClose,
  onCommentAdded,
}: CommentSheetProps) {
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!isOpen || !postId) return;
    let isSubscribed = true;

    api
      .get<Comment[]>(API.POST_COMMENTS(postId))
      .then((res) => {
        if (!isSubscribed) return;
        if (res.success && res.data) {
          setComments(res.data);
        }
        setLoading(false);
      })
      .catch((error: unknown) => {
        if (!isSubscribed) return;
        console.error('Failed to fetch comments:', error);
        setLoading(false);
      });

    return () => {
      isSubscribed = false;
    };
  }, [isOpen, postId]);

  const handleSubmit = async () => {
    if (!content.trim() || isSubmitting) return;

    setIsSubmitting(true);
    try {
      const res = await api.post<Comment>(API.POST_COMMENTS(postId), { content });
      if (res.success && res.data) {
        setComments((prev) => [res.data, ...prev]);
        setContent('');
        if (onCommentAdded) onCommentAdded();
      }
    } catch (error: unknown) {
      console.error('Failed to post comment:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className={styles.overlay}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className={styles.sheet}>
        <div className={styles.header}>
          <h2>Comments</h2>
          <button
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Close comments"
          >
            <svg
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <div className={styles.content}>
          {loading ? (
            <div className={styles.loading}>Loading comments...</div>
          ) : comments.length === 0 ? (
            <div className={styles.empty}>
              No comments yet. Be the first to share your thoughts!
            </div>
          ) : (
            <div className={styles.commentList}>
              {comments.map((comment) => (
                <div key={comment.id} className={styles.comment}>
                  {comment.user.avatar_url ? (
                    <img
                      src={comment.user.avatar_url}
                      alt={comment.user.display_name}
                      className={styles.avatar}
                    />
                  ) : (
                    <div className={styles.avatarFallback}>
                      {comment.user.display_name?.[0] || '?'}
                    </div>
                  )}
                  <div className={styles.commentBody}>
                    <div className={styles.authorRow}>
                      <span className={styles.displayName}>
                        {comment.user.display_name}
                      </span>
                      <span className={styles.username}>
                        @{comment.user.username}
                      </span>
                      <span className={styles.timestamp}>
                        {relativeTime(comment.created_at)}
                      </span>
                    </div>
                    <div className={styles.text}>{comment.content}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className={styles.footer}>
          <div className={styles.inputWrapper}>
            <textarea
              className={styles.textarea}
              placeholder="Post your reply..."
              value={content}
              onChange={(e) => setContent(e.target.value)}
              disabled={isSubmitting}
            />
          </div>
          <Button
            variant="primary"
            size="sm"
            loading={isSubmitting}
            disabled={!content.trim()}
            onClick={() => {
              void handleSubmit();
            }}
          >
            Reply
          </Button>
        </div>
      </div>
    </div>
  );
}
