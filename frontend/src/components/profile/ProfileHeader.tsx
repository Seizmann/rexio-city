'use client';

import React, { useState } from 'react';
import styles from './ProfileHeader.module.css';
import type { User, FollowCounts } from '@/lib/types';
import Button from '@/components/ui/Button';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import EditProfileModal from './EditProfileModal';

interface ProfileHeaderProps {
  user: User;
  followCounts: FollowCounts;
  isFollowing: boolean;
  isOwnProfile: boolean;
  onFollowChange?: (following: boolean) => void;
  onEditProfile?: () => void;
}

export default function ProfileHeader({
  user,
  followCounts,
  isFollowing: initialIsFollowing,
  isOwnProfile,
  onFollowChange,
  onEditProfile,
}: ProfileHeaderProps) {
  const [isFollowing, setIsFollowing] = useState(initialIsFollowing);
  const [loading, setLoading] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [localCounts, setLocalCounts] = useState(followCounts);

  const followerCount =
    localCounts.follower_count ??
    localCounts.followers ??
    user.follower_count ??
    user.followers ??
    0;

  const followingCount =
    localCounts.following_count ??
    localCounts.following ??
    user.following_count ??
    user.following ??
    0;

  const handleFollowToggle = async () => {
    if (loading) return;
    setLoading(true);

    // Optimistic update
    const previousState = isFollowing;
    const previousCounts = { ...localCounts };
    const nextFollowerCount = previousState
      ? Math.max(0, followerCount - 1)
      : followerCount + 1;

    setIsFollowing(!previousState);
    setLocalCounts({
      ...localCounts,
      followers: nextFollowerCount,
      follower_count: nextFollowerCount,
    });

    if (onFollowChange) {
      onFollowChange(!previousState);
    }

    try {
      if (previousState) {
        await api.delete(API.FOLLOW(user.id));
      } else {
        await api.post(API.FOLLOW(user.id));
      }
    } catch {
      // Revert on error
      setIsFollowing(previousState);
      setLocalCounts(previousCounts);
      if (onFollowChange) {
        onFollowChange(previousState);
      }
    } finally {
      setLoading(false);
    }
  };

  const getInitials = (name: string) => {
    return name ? name.charAt(0).toUpperCase() : '?';
  };

  const handleSaveProfile = () => {
    setIsEditModalOpen(false);
    if (onEditProfile) {
      onEditProfile(); // Trigger parent refresh if needed
    }
  };

  return (
    <div className={styles.header}>
      <div
        className={`${styles.cover} ${isOwnProfile ? styles.coverEditable : ''}`}
        style={
          user.cover_url
            ? { backgroundImage: `url(${user.cover_url})` }
            : undefined
        }
        onClick={() => {
          if (isOwnProfile) setIsEditModalOpen(true);
        }}
      >
        {isOwnProfile && (
          <div className={styles.coverOverlay}>
            <div className={styles.editIconBtn} aria-label="Edit cover photo">
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
                <circle cx="12" cy="13" r="4" />
              </svg>
            </div>
          </div>
        )}
      </div>

      <div className={styles.infoContainer}>
        <div className={styles.avatarActionRow}>
          <div
            className={`${styles.avatar} ${isOwnProfile ? styles.avatarEditable : ''}`}
            onClick={() => {
              if (isOwnProfile) setIsEditModalOpen(true);
            }}
          >
            {user.avatar_url ? (
              <img
                src={user.avatar_url}
                alt={user.display_name || user.username}
                className={styles.avatarImage}
              />
            ) : (
              <span className={styles.avatarInitials}>
                {getInitials(user.display_name || user.username)}
              </span>
            )}

            {isOwnProfile && (
              <div className={styles.avatarOverlay}>
                <div className={styles.editIconBtn} aria-label="Edit avatar photo">
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
                    <circle cx="12" cy="13" r="4" />
                  </svg>
                </div>
              </div>
            )}
          </div>

          <div className={styles.actionButton}>
            {isOwnProfile ? (
              <Button
                variant="secondary"
                onClick={() => setIsEditModalOpen(true)}
              >
                Edit Profile
              </Button>
            ) : (
              <Button
                variant={isFollowing ? 'secondary' : 'primary'}
                onClick={() => {
                  void handleFollowToggle();
                }}
                loading={loading}
              >
                {isFollowing ? 'Following' : 'Follow'}
              </Button>
            )}
          </div>
        </div>

        <div className={styles.details}>
          <div className={styles.nameRow}>
            <h1 className={styles.displayName}>
              {user.display_name || user.username}
            </h1>
            <p className={styles.username}>@{user.username}</p>
          </div>

          {user.bio && <p className={styles.bio}>{user.bio}</p>}

          <div className={styles.followCounts}>
            <span className={styles.countItem}>
              <span className={styles.countValue}>{followingCount}</span>{' '}
              Following
            </span>
            <span className={styles.countItem}>
              <span className={styles.countValue}>{followerCount}</span>{' '}
              Followers
            </span>
          </div>
        </div>
      </div>

      <EditProfileModal
        user={user}
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        onSave={handleSaveProfile}
      />
    </div>
  );
}
