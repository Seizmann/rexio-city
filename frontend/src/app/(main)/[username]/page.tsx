'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import { useCachedFetch } from '@/lib/useCachedFetch';
import type { User, Post, FollowCounts } from '@/lib/types';
import { useAuth } from '@/context/AuthContext';
import ProfileHeader from '@/components/profile/ProfileHeader';
import ProfileTabs from '@/components/profile/ProfileTabs';
import PostCard from '@/components/feed/PostCard';
import { PostCardSkeleton } from '@/components/ui/Skeleton';
import styles from './page.module.css';

export default function ProfilePage() {
  const params = useParams();
  const rawUsername = params.username as string;
  const username = rawUsername?.startsWith('@') ? rawUsername.slice(1) : rawUsername;
  const { user: authUser } = useAuth();

  const [activeTab, setActiveTab] = useState('posts');
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [isFollowing, setIsFollowing] = useState(false);

  // Use cached fetch for profile user data
  const { data: user, loading: userLoading } = useCachedFetch<User | null>(
    `profile-${username}`,
    async () => {
      const res = await api.get<User>(API.USER(username));
      return res.success && res.data ? res.data : null;
    },
    {
      dependencies: [username, refreshTrigger],
      onSuccess: (userData) => {
        // Use is_following from user object if available
        if (userData && userData.is_following !== undefined) {
          setIsFollowing(userData.is_following);
        }
      },
    },
  );
  // Use cached fetch for user posts (only when user profile is loaded)
  const postsKey = user ? `posts-user-${user.id}` : '';
  const { data: posts, loading: postsLoading } = useCachedFetch<Post[]>(
    postsKey,
    async () => {
      if (!user) return [];
      const res = await api.get<Post[]>(`${API.POSTS}?user_id=${user.id}`);
      return res.success && res.data ? res.data : [];
    },
    {
      dependencies: [user?.id, refreshTrigger],
    },
  );
  // Use follow counts from user object (already includes follower_count and following_count)
  const followCounts: FollowCounts = {
    follower_count: user?.follower_count ?? user?.followers ?? 0,
    following_count: user?.following_count ?? user?.following ?? 0,
    followers: user?.followers ?? user?.follower_count ?? 0,
    following: user?.following ?? user?.following_count ?? 0,
  };

  const loading = userLoading || postsLoading;

  function handleEditProfile() {
    setRefreshTrigger((prev) => prev + 1);
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <PostCardSkeleton />
        <PostCardSkeleton />
      </div>
    );
  }

  if (!user) {
    return (
      <div className={styles.errorState}>
        <h2>Profile Not Found</h2>
        <p>This account does not exist.</p>
      </div>
    );
  }

  const isOwnProfile = authUser?.id === user.id;

  return (
    <div className={styles.container}>
      <ProfileHeader
        user={user}
        followCounts={followCounts}
        isFollowing={isFollowing}
        isOwnProfile={isOwnProfile}
        onEditProfile={handleEditProfile}
      />

      <ProfileTabs activeTab={activeTab} onTabChange={setActiveTab} />

      <div className={styles.content}>
        {activeTab === 'posts' && (
          <div className={styles.feed}>
            {posts && posts.length > 0 ? (
              posts.map((post) => <PostCard key={post.id} post={post} />)
            ) : (
              <div className={styles.emptyState}>
                <p>No posts yet.</p>
                <p className={styles.emptySub}>
                  When @{user.username} posts, it will show up here.
                </p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'replies' && (
          <div className={styles.emptyState}>
            <p>Replies will appear here in a future update.</p>
          </div>
        )}

        {activeTab === 'media' && (
          <div className={styles.feed}>
            {posts && posts.filter((p) => p.media && p.media.length > 0).length > 0 ? (
              posts
                .filter((p) => p.media && p.media.length > 0)
                .map((post) => <PostCard key={post.id} post={post} />)
            ) : (
              <div className={styles.emptyState}>
                <p>No media posts yet.</p>
                <p className={styles.emptySub}>
                  When @{user.username} posts photos or videos, they will show up here.
                </p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
