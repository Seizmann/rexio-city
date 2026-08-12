'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { api } from '@/lib/api';
import { API } from '@/lib/constants';
import type { User, Post, FollowCounts } from '@/lib/types';
import { useAuth } from '@/context/AuthContext';
import ProfileHeader from '@/components/profile/ProfileHeader';
import ProfileTabs from '@/components/profile/ProfileTabs';
import PostCard from '@/components/feed/PostCard';
import { PostCardSkeleton } from '@/components/ui/Skeleton';
import styles from './page.module.css';

export default function ProfilePage() {
  const params = useParams();
  const username = params.username as string;
  const { user: authUser } = useAuth();

  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [followCounts, setFollowCounts] = useState<FollowCounts>({
    follower_count: 0,
    following_count: 0,
  });
  const [isFollowing, setIsFollowing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('posts');
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  useEffect(() => {
    if (!username) return;
    let isSubscribed = true;

    api
      .get<User>(API.USER(username))
      .then(async (userRes) => {
        if (!isSubscribed) return;
        if (!userRes.success || !userRes.data) {
          setError('This account does not exist.');
          setLoading(false);
          return;
        }

        const profileUser = userRes.data;
        setUser(profileUser);

        const [countsRes, followingRes, postsRes] = await Promise.all([
          api.get<FollowCounts>(API.FOLLOW_COUNTS(profileUser.id)),
          authUser
            ? api.get<{ is_following: boolean }>(
                API.IS_FOLLOWING(profileUser.id),
              )
            : Promise.resolve({
                success: true,
                data: { is_following: false },
              }),
          api.get<Post[]>(`${API.POSTS}?user_id=${profileUser.id}`),
        ]);

        if (!isSubscribed) return;
        if (countsRes.success && countsRes.data) {
          setFollowCounts(countsRes.data);
        }
        if (followingRes.success && followingRes.data) {
          setIsFollowing(followingRes.data.is_following);
        }
        if (postsRes.success && postsRes.data) {
          setPosts(postsRes.data);
        }
        setLoading(false);
      })
      .catch((err: unknown) => {
        if (!isSubscribed) return;
        console.error('Error fetching profile:', err);
        setError('An error occurred while loading the profile.');
        setLoading(false);
      });

    return () => {
      isSubscribed = false;
    };
  }, [username, authUser, refreshTrigger]);

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

  if (error || !user) {
    return (
      <div className={styles.errorState}>
        <h2>Profile Not Found</h2>
        <p>{error || 'This account does not exist.'}</p>
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
            {posts.length > 0 ? (
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
          <div className={styles.emptyState}>
            <p>Media posts will appear here in a future update.</p>
          </div>
        )}
      </div>
    </div>
  );
}
