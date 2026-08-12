'use client';

import React, { useState } from 'react';
import styles from './PostComposer.module.css';
import { useAuth } from '@/context/AuthContext';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import type { Post } from '@/lib/types';
import Button from '@/components/ui/Button';

interface PostComposerProps {
  onPostCreated: (post: Post) => void;
}

const MAX_CHARS = 500;

export default function PostComposer({ onPostCreated }: PostComposerProps) {
  const { user } = useAuth();
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const remainingChars = MAX_CHARS - content.length;
  const isOverLimit = remainingChars < 0;
  const isWarning = remainingChars <= 20;
  const isEmpty = content.trim().length === 0;

  const handleSubmit = async () => {
    if (isEmpty || isOverLimit || isSubmitting) return;

    setIsSubmitting(true);
    try {
      const res = await api.post<Post>(API.POSTS, { content });
      if (res.success && res.data) {
        onPostCreated(res.data);
        setContent('');
      }
    } catch (error) {
      console.error('Failed to post:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={styles.composer}>
      {user?.avatar_url ? (
        <img src={user.avatar_url} alt={user.display_name} className={styles.avatar} />
      ) : (
        <div className={styles.avatarFallback}>
          {user?.display_name?.[0] || '?'}
        </div>
      )}
      
      <div className={styles.content}>
        <textarea
          className={styles.textarea}
          placeholder="What's happening?"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          disabled={isSubmitting}
        />
        
        <div className={styles.footer}>
          <span className={`${styles.counter} ${(isWarning || isOverLimit) ? styles.counterError : ''}`}>
            {remainingChars}
          </span>
          <div className={styles.actions}>
            <Button 
              variant="primary" 
              size="sm"
              loading={isSubmitting}
              disabled={isEmpty || isOverLimit}
              onClick={() => { void handleSubmit(); }}
            >
              Post
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
