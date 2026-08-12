'use client';

import React, { useState, useEffect, useRef } from 'react';
import styles from './EditProfileModal.module.css';
import type { User } from '@/lib/types';
import { api } from '@/lib/api';
import { API, LIMITS } from '@/lib/constants';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import { useAuth } from '@/context/AuthContext';

interface EditProfileModalProps {
  user: User;
  isOpen: boolean;
  onClose: () => void;
  onSave: (user: User) => void;
}

// Inner form component — remounted via key when modal opens, so we can use
// the initial prop values directly as useState defaults without setState-in-effect.
function EditProfileForm({
  user,
  onClose,
  onSave,
}: {
  user: User;
  onClose: () => void;
  onSave: (user: User) => void;
}) {
  const { setUser: setAuthUser } = useAuth();
  const [displayName, setDisplayName] = useState(user.display_name || '');
  const [bio, setBio] = useState(user.bio || '');

  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(user.avatar_url || null);
  const [coverFile, setCoverFile] = useState<File | null>(null);
  const [coverPreview, setCoverPreview] = useState<string | null>(user.cover_url || null);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const avatarInputRef = useRef<HTMLInputElement>(null);
  const coverInputRef = useRef<HTMLInputElement>(null);

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (!file.type.startsWith('image/')) {
      setError('Please select an image file (PNG, JPG, WebP).');
      return;
    }
    setAvatarFile(file);
    setAvatarPreview(URL.createObjectURL(file));
  };

  const handleCoverChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (!file.type.startsWith('image/')) {
      setError('Please select an image file (PNG, JPG, WebP).');
      return;
    }
    setCoverFile(file);
    setCoverPreview(URL.createObjectURL(file));
  };

  interface UploadResponse {
    success: boolean;
    data?: { url?: string };
    error?: { message?: string };
  }

  const uploadFile = async (file: File): Promise<string> => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch(API.MEDIA_UPLOAD, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error('Failed to upload image file');
    }

    const data = (await response.json()) as UploadResponse;
    if (!data.success || !data.data?.url) {
      throw new Error(data.error?.message ?? 'Failed to upload image file');
    }

    return data.data.url;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      let finalAvatarUrl = user.avatar_url;
      let finalCoverUrl = user.cover_url;

      if (avatarFile) {
        finalAvatarUrl = await uploadFile(avatarFile);
      }
      if (coverFile) {
        finalCoverUrl = await uploadFile(coverFile);
      }

      const res = await api.patch<User>(API.USERS_ME, {
        display_name: displayName,
        bio: bio,
        avatar_url: finalAvatarUrl,
        cover_url: finalCoverUrl,
      });

      if (res.success && res.data) {
        setAuthUser(res.data);
        onSave(res.data);
      } else {
        setError(res.error?.message ?? 'Failed to update profile');
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'An error occurred';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form className={styles.form} onSubmit={(e) => void handleSubmit(e)}>
      {/* Media Upload Section */}
      <div className={styles.mediaSection}>
        <div
          className={styles.coverContainer}
          style={coverPreview ? { backgroundImage: `url(${coverPreview})` } : undefined}
        >
          <div className={styles.coverOverlay}>
            <button
              type="button"
              className={styles.uploadIconBtn}
              onClick={() => coverInputRef.current?.click()}
              aria-label="Add cover photo"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
                <circle cx="12" cy="13" r="4" />
              </svg>
            </button>
          </div>
          <input ref={coverInputRef} type="file" accept="image/*" className={styles.hiddenInput} onChange={handleCoverChange} />
        </div>

        <div className={styles.avatarContainer}>
          {avatarPreview ? (
            <img src={avatarPreview} alt={user.display_name ?? user.username} className={styles.avatarImg} />
          ) : (
            <div className={styles.avatarFallback}>
              {user.display_name?.[0] ?? user.username?.[0] ?? '?'}
            </div>
          )}
          <div className={styles.avatarOverlay}>
            <button
              type="button"
              className={styles.uploadIconBtn}
              onClick={() => avatarInputRef.current?.click()}
              aria-label="Add profile photo"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
                <circle cx="12" cy="13" r="4" />
              </svg>
            </button>
          </div>
          <input ref={avatarInputRef} type="file" accept="image/*" className={styles.hiddenInput} onChange={handleAvatarChange} />
        </div>
      </div>

      {error && <div className={styles.errorBanner}>{error}</div>}

      <div className={styles.fieldsContainer}>
        <Input
          label="Display Name"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          maxLength={LIMITS.DISPLAY_NAME_MAX_CHARS}
          disabled={loading}
        />

        <div className={styles.field}>
          <label htmlFor="bio" className={styles.label}>Bio</label>
          <textarea
            id="bio"
            className={styles.textarea}
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            maxLength={LIMITS.BIO_MAX_CHARS}
            disabled={loading}
          />
        </div>
      </div>

      <div className={styles.footer}>
        <Button variant="ghost" onClick={onClose} disabled={loading} type="button">
          Cancel
        </Button>
        <Button variant="primary" type="submit" loading={loading}>
          Save changes
        </Button>
      </div>
    </form>
  );
}

export default function EditProfileModal({ user, isOpen, onClose, onSave }: EditProfileModalProps) {
  // Escape key to close
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) onClose();
  };

  if (!isOpen) return null;

  return (
    <div className={styles.overlay} onClick={handleOverlayClick}>
      <div className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="edit-profile-title">
        <div className={styles.header}>
          <h2 id="edit-profile-title" className={styles.title}>Edit Profile</h2>
          <button className={styles.closeButton} onClick={onClose} aria-label="Close modal">&times;</button>
        </div>
        {/* key=user.id forces a fresh mount each time the modal opens,
            so the inner form re-initialises from current user props */}
        <EditProfileForm key={String(user.id) + String(isOpen)} user={user} onClose={onClose} onSave={onSave} />
      </div>
    </div>
  );
}
