'use client';

import React, { useState, useEffect } from 'react';
import styles from './EditProfileModal.module.css';
import type { User } from '@/lib/types';
import { api } from '@/lib/api';
import { API, LIMITS } from '@/lib/constants';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';

interface EditProfileModalProps {
  user: User;
  isOpen: boolean;
  onClose: () => void;
  onSave: (user: User) => void;
}

export default function EditProfileModal({
  user,
  isOpen,
  onClose,
  onSave,
}: EditProfileModalProps) {
  const [displayName, setDisplayName] = useState(user.display_name || '');
  const [bio, setBio] = useState(user.bio || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      window.addEventListener('keydown', handleKeyDown);
    }
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await api.patch<User>(API.USERS_ME, {
        display_name: displayName,
        bio: bio,
      });

      if (res.success && res.data) {
        onSave(res.data);
      } else {
        setError(res.error?.message || 'Failed to update profile');
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'An error occurred';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className={styles.overlay} onClick={handleOverlayClick}>
      <div
        className={styles.modal}
        role="dialog"
        aria-modal="true"
        aria-labelledby="edit-profile-title"
      >
        <div className={styles.header}>
          <h2 id="edit-profile-title" className={styles.title}>
            Edit Profile
          </h2>
          <button
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Close modal"
          >
            &times;
          </button>
        </div>
        <form className={styles.form} onSubmit={(e) => void handleSubmit(e)}>
          {error && <div style={{ color: 'var(--error)' }}>{error}</div>}

          <Input
            label="Display Name"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            maxLength={LIMITS.DISPLAY_NAME_MAX_CHARS}
            disabled={loading}
          />

          <div className={styles.field}>
            <label htmlFor="bio" className={styles.label}>
              Bio
            </label>
            <textarea
              id="bio"
              className={styles.textarea}
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              maxLength={LIMITS.BIO_MAX_CHARS}
              disabled={loading}
            />
          </div>

          <div className={styles.footer}>
            <Button
              variant="ghost"
              onClick={onClose}
              disabled={loading}
              type="button"
            >
              Cancel
            </Button>
            <Button variant="primary" type="submit" loading={loading}>
              Save changes
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
