/**
 * Shared TypeScript types for RexiO City frontend.
 * Must stay in sync with Go backend models (internal/models/models.go).
 */

/* ── User ─────────────────────────────────────────────────────── */

export interface User {
  id: number;
  username: string;
  display_name: string;
  email?: string; // only present on own profile
  bio?: string;
  avatar_url?: string;
  cover_url?: string;
  followers?: number;
  following?: number;
  follower_count?: number;
  following_count?: number;
  is_following?: boolean; // computed by backend, not stored in DB
  created_at?: string;
  updated_at?: string;
}

/* ── Posts ─────────────────────────────────────────────────────── */

export interface Post {
  id: number;
  public_id?: string;
  user_id: number;
  user?: User;
  content: string;
  media?: PostMedia[];
  like_count?: number;
  likes?: number;
  comment_count?: number;
  comments?: number;
  repost_count?: number;
  reposts?: number;
  bookmark_count?: number;
  is_liked?: boolean;
  is_reposted?: boolean;
  is_bookmarked?: boolean;
  created_at: string;
  // Optimistic upload fields — only present on pending posts in the feed.
  // Cleared once the real post is confirmed from the server.
  _pending?: true;
  _uploadStatus?: 'uploading' | 'updating' | 'finishing' | 'done' | 'error';
  _localPreviews?: { previewUrl: string; type: 'photo' | 'video' }[];
  _pendingKey?: string; // unique key to identify this pending post
}

export interface PostMedia {
  id?: number;
  media_url: string;
  media_type: string; // photo, video, voice
  order?: number;
  order_index?: number;
}

export interface Comment {
  id: number;
  user_id: number;
  user?: User;
  post_id: number;
  parent_id: number | null;
  content: string;
  created_at: string;
}

/* ── Auth ──────────────────────────────────────────────────────── */

export interface AuthData {
  user: User;
  access_token: string;
  // refresh_token intentionally absent: it is now set as an httpOnly cookie
  // by the backend and is never accessible to JavaScript.
  expires_in: number;
}

/* ── Session (device management) ──────────────────────────────── */

export interface Session {
  id: number;
  user_id: number;
  parent_session_id: number | null;
  device_info: string;    // User-Agent string
  ip_address: string;
  created_at: string;
  last_used_at: string;
  expires_at: string;
  revoked_at: string | null; // null = active
}

export interface SignupPayload {
  username: string;
  email: string;
  password: string;
  display_name: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

/* ── Follow ────────────────────────────────────────────────────── */

export interface FollowCounts {
  follower_count?: number;
  following_count?: number;
  followers?: number;
  following?: number;
}

/* ── Search ─────────────────────────────────────────────────────── */

export interface SearchUserResult {
  id: number;
  username: string;
  display_name?: string;
  avatar_url?: string;
}

export interface SearchPostResult {
  id: number;
  public_id: string;
  content: string;
  user_id: number;
  created_at: string;
  user: SearchUserResult;
}

export interface SearchHashtagResult {
  hashtag: string;
  count: number;
  posts?: SearchPostResult[];
}

export interface SearchResponse {
  users?: SearchUserResult[];
  posts?: SearchPostResult[];
  hashtags?: SearchHashtagResult[];
  total: number;
  has_users: boolean;
  has_posts: boolean;
  has_hashtags: boolean;
}

export interface APIResponse<T> {
  success: boolean;
  data: T;
  meta?: {
    page: number;
    per_page: number;
    total: number;
  };
  error?: APIError;
}

export interface APIError {
  code: string;
  message: string;
  details?: Record<string, string>;
}
