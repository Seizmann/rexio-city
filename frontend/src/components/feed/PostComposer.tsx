'use client';

import React, { useState, useRef } from 'react';
import styles from './PostComposer.module.css';
import { useAuth } from '@/context/AuthContext';
import Button from '@/components/ui/Button';

interface PostComposerProps {
  onPostSubmit: (payload: {
    content: string;
    files: { file: File; type: 'photo' | 'video' }[];
    pendingKey: string;
  }) => void;
}

// Local preview — no upload happens here, file stays in browser memory
interface LocalAttachment {
  file: File;
  type: 'photo' | 'video';
  previewUrl: string;
}

const MAX_CHARS = 500;
const MAX_PHOTOS = 10;

export default function PostComposer({ onPostSubmit }: PostComposerProps) {
  const { user } = useAuth();
  const [content, setContent] = useState('');
  const [attachments, setAttachments] = useState<LocalAttachment[]>([]);
  const [attachError, setAttachError] = useState<string | null>(null);

  const imageInputRef = useRef<HTMLInputElement>(null);
  const videoInputRef = useRef<HTMLInputElement>(null);

  const remainingChars = MAX_CHARS - content.length;
  const isOverLimit = remainingChars < 0;
  const isWarning = remainingChars <= 20;
  const isContentEmpty = content.trim().length === 0;
  const hasAttachments = attachments.length > 0;
  const isSubmitDisabled = (isContentEmpty && !hasAttachments) || isOverLimit;

  const handleFileSelect = (files: FileList | null, type: 'photo' | 'video') => {
    if (!files || files.length === 0) return;
    setAttachError(null);

    const fileArray = Array.from(files);
    const currentVideos = attachments.filter((a) => a.type === 'video');
    const currentPhotos = attachments.filter((a) => a.type === 'photo');

    if (type === 'video') {
      if (currentPhotos.length > 0) {
        setAttachError('Cannot attach video when photos are already attached.');
        return;
      }
      if (currentVideos.length >= 1 || fileArray.length > 1) {
        setAttachError('You can attach max 1 video per post.');
        return;
      }
    } else {
      if (currentVideos.length > 0) {
        setAttachError('Cannot attach photos when a video is already attached.');
        return;
      }
      if (currentPhotos.length + fileArray.length > MAX_PHOTOS) {
        setAttachError(`You can attach up to ${MAX_PHOTOS} photos per post.`);
        return;
      }
    }

    const newAttachments: LocalAttachment[] = fileArray.map((file) => ({
      file,
      type,
      previewUrl: URL.createObjectURL(file),
    }));

    setAttachments((prev) => [...prev, ...newAttachments]);

    if (imageInputRef.current) imageInputRef.current.value = '';
    if (videoInputRef.current) videoInputRef.current.value = '';
  };

  const handleRemove = (idx: number) => {
    setAttachments((prev) => {
      URL.revokeObjectURL(prev[idx].previewUrl);
      return prev.filter((_, i) => i !== idx);
    });
  };

  const handleSubmit = () => {
    if (isSubmitDisabled) return;

    // Generate a stable key so the feed can match this pending card
    const pendingKey = `pending_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`;

    onPostSubmit({
      content,
      files: attachments.map((a) => ({ file: a.file, type: a.type })),
      pendingKey,
    });

    // Clear the composer immediately — upload happens in the parent
    setContent('');
    // Revoke all object URLs before clearing
    attachments.forEach((a) => URL.revokeObjectURL(a.previewUrl));
    setAttachments([]);
    setAttachError(null);
  };

  return (
    <div className={styles.composer}>
      {user?.avatar_url ? (
        <img src={user.avatar_url} alt={user.display_name} className={styles.avatar} />
      ) : (
        <div className={styles.avatarFallback}>{user?.display_name?.[0] || '?'}</div>
      )}

      <div className={styles.content}>
        <textarea
          className={styles.textarea}
          placeholder="What's happening?"
          value={content}
          onChange={(e) => setContent(e.target.value)}
        />

        {/* Local preview grid — shown before upload */}
        {attachments.length > 0 && (
          <div className={styles.mediaPreviewGrid}>
            {attachments.map((item, idx) => (
              <div key={idx} className={styles.mediaPreviewItem}>
                {item.type === 'video' ? (
                  <video src={item.previewUrl} className={styles.previewVideo} muted />
                ) : (
                  <img src={item.previewUrl} alt="Preview" className={styles.previewImage} />
                )}
                <button
                  type="button"
                  className={styles.removeBtn}
                  onClick={() => handleRemove(idx)}
                  aria-label="Remove attachment"
                >
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2">
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                </button>
              </div>
            ))}
          </div>
        )}

        {attachError && <div className={styles.errorMessage}>{attachError}</div>}

        <div className={styles.footer}>
          <div className={styles.toolbar}>
            {/* Hidden file inputs */}
            <input
              type="file"
              ref={imageInputRef}
              accept="image/*"
              multiple
              style={{ display: 'none' }}
              onChange={(e) => handleFileSelect(e.target.files, 'photo')}
            />
            <input
              type="file"
              ref={videoInputRef}
              accept="video/*"
              style={{ display: 'none' }}
              onChange={(e) => handleFileSelect(e.target.files, 'video')}
            />

            {/* Photo Attachment Button */}
            <button
              type="button"
              className={styles.toolBtn}
              onClick={() => imageInputRef.current?.click()}
              title="Attach photos (up to 10)"
              disabled={attachments.some((a) => a.type === 'video')}
              aria-label="Attach photos"
            >
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <polyline points="21 15 16 10 5 21" />
              </svg>
            </button>

            {/* Video Attachment Button */}
            <button
              type="button"
              className={styles.toolBtn}
              onClick={() => videoInputRef.current?.click()}
              title="Attach video (max 500MB)"
              disabled={attachments.length > 0}
              aria-label="Attach video"
            >
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="23 7 16 12 23 17 23 7" />
                <rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
              </svg>
            </button>
          </div>

          <div className={styles.submitRow}>
            <span className={`${styles.counter} ${isWarning || isOverLimit ? styles.counterError : ''}`}>
              {remainingChars}
            </span>
            <Button
              variant="primary"
              size="sm"
              disabled={isSubmitDisabled}
              onClick={handleSubmit}
            >
              Post
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
