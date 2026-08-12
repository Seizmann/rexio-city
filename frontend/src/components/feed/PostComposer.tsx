'use client';

import React, { useState, useRef } from 'react';
import styles from './PostComposer.module.css';
import { useAuth } from '@/context/AuthContext';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import type { Post } from '@/lib/types';
import Button from '@/components/ui/Button';

interface PostComposerProps {
  onPostCreated: (post: Post) => void;
}

interface AttachedMediaItem {
  id: string;
  url: string;
  type: 'photo' | 'video' | 'voice';
  previewUrl: string;
  isUploading: boolean;
  error?: string;
}

const MAX_CHARS = 500;
const MAX_PHOTOS = 10;

export default function PostComposer({ onPostCreated }: PostComposerProps) {
  const { user } = useAuth();
  const [content, setContent] = useState('');
  const [mediaList, setMediaList] = useState<AttachedMediaItem[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

  const imageInputRef = useRef<HTMLInputElement>(null);
  const videoInputRef = useRef<HTMLInputElement>(null);

  const remainingChars = MAX_CHARS - content.length;
  const isOverLimit = remainingChars < 0;
  const isWarning = remainingChars <= 20;
  const isAnyUploading = mediaList.some((item) => item.isUploading);
  const hasMedia = mediaList.length > 0 && mediaList.some((item) => !item.isUploading && !!item.url);
  const isContentEmpty = content.trim().length === 0;

  // Disabled if (no content AND no uploaded media), or over char limit, or currently uploading, or submitting
  const isSubmitDisabled = (isContentEmpty && !hasMedia) || isOverLimit || isAnyUploading || isSubmitting;

  const handleFileUpload = async (files: FileList | null, expectedType: 'photo' | 'video') => {
    if (!files || files.length === 0) return;
    setUploadError(null);

    const fileArray = Array.from(files);

    const currentVideos = mediaList.filter((m) => m.type === 'video');
    const currentPhotos = mediaList.filter((m) => m.type === 'photo');

    if (expectedType === 'video') {
      if (currentPhotos.length > 0) {
        setUploadError('Cannot attach video when photos are already attached.');
        return;
      }
      if (currentVideos.length >= 1 || fileArray.length > 1) {
        setUploadError('You can attach max 1 video per post.');
        return;
      }
    } else {
      if (currentVideos.length > 0) {
        setUploadError('Cannot attach photos when a video is already attached.');
        return;
      }
      if (currentPhotos.length + fileArray.length > MAX_PHOTOS) {
        setUploadError(`You can attach up to ${MAX_PHOTOS} photos per post.`);
        return;
      }
    }

    for (const file of fileArray) {
      const tempId = `${Date.now()}_${Math.random().toString(36).substring(2, 7)}`;
      const previewUrl = URL.createObjectURL(file);

      const newItem: AttachedMediaItem = {
        id: tempId,
        url: '',
        type: expectedType,
        previewUrl,
        isUploading: true,
      };

      setMediaList((prev) => [...prev, newItem]);

      const formData = new FormData();
      formData.append('file', file);

      try {
        const res = await api.upload<{ url: string; type: string }>(API.MEDIA_UPLOAD, formData);
        if (res.success && res.data?.url) {
          const uploadedUrl = res.data.url;
          const mediaType = (res.data.type || expectedType) as 'photo' | 'video' | 'voice';

          setMediaList((prev) =>
            prev.map((item) =>
              item.id === tempId
                ? { ...item, url: uploadedUrl, type: mediaType, isUploading: false }
                : item
            )
          );
        } else {
          setUploadError(res.error?.message || 'Failed to upload media file.');
          setMediaList((prev) => prev.filter((item) => item.id !== tempId));
        }
      } catch (err) {
        console.error('Media upload error:', err);
        setUploadError('Upload failed. Please check network connection.');
        setMediaList((prev) => prev.filter((item) => item.id !== tempId));
      }
    }

    // Reset file input values
    if (imageInputRef.current) imageInputRef.current.value = '';
    if (videoInputRef.current) videoInputRef.current.value = '';
  };

  const handleRemoveMedia = (id: string) => {
    setMediaList((prev) => prev.filter((item) => item.id !== id));
  };

  const handleSubmit = async () => {
    if (isSubmitDisabled) return;

    setIsSubmitting(true);
    setUploadError(null);

    const validMedia = mediaList.filter((item) => !item.isUploading && !!item.url);
    const mediaUrls = validMedia.map((item) => item.url);
    const mediaTypes = validMedia.map((item) => item.type);

    try {
      const res = await api.post<Post>(API.POSTS, {
        content,
        media_urls: mediaUrls,
        media_types: mediaTypes,
      });

      if (res.success && res.data) {
        onPostCreated(res.data);
        setContent('');
        setMediaList([]);
      } else {
        setUploadError(res.error?.message || 'Failed to create post.');
      }
    } catch (error) {
      console.error('Failed to post:', error);
      setUploadError('An error occurred while publishing post.');
    } finally {
      setIsSubmitting(false);
    }
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
          disabled={isSubmitting}
        />

        {/* Media Previews */}
        {mediaList.length > 0 && (
          <div className={styles.mediaPreviewGrid}>
            {mediaList.map((item) => (
              <div key={item.id} className={styles.mediaPreviewItem}>
                {item.type === 'video' ? (
                  <video src={item.previewUrl} className={styles.previewVideo} muted />
                ) : (
                  <img src={item.previewUrl} alt="Preview" className={styles.previewImage} />
                )}

                {item.isUploading && (
                  <div className={styles.uploadOverlay}>
                    <div className={styles.spinner} />
                  </div>
                )}

                <button
                  type="button"
                  className={styles.removeBtn}
                  onClick={() => handleRemoveMedia(item.id)}
                  aria-label="Remove media"
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

        {uploadError && <div className={styles.errorMessage}>{uploadError}</div>}

        <div className={styles.footer}>
          <div className={styles.toolbar}>
            {/* Hidden file inputs */}
            <input
              type="file"
              ref={imageInputRef}
              accept="image/*"
              multiple
              style={{ display: 'none' }}
              onChange={(e) => {
                void handleFileUpload(e.target.files, 'photo');
              }}
            />
            <input
              type="file"
              ref={videoInputRef}
              accept="video/*"
              style={{ display: 'none' }}
              onChange={(e) => {
                void handleFileUpload(e.target.files, 'video');
              }}
            />

            {/* Photo Attachment Button */}
            <button
              type="button"
              className={styles.toolBtn}
              onClick={() => imageInputRef.current?.click()}
              title="Attach photos (up to 10)"
              disabled={isSubmitting || mediaList.some((m) => m.type === 'video')}
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
              disabled={isSubmitting || mediaList.length > 0}
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
              loading={isSubmitting || isAnyUploading}
              disabled={isSubmitDisabled}
              onClick={() => {
                void handleSubmit();
              }}
            >
              Post
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
