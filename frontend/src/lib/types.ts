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
  bio: string;
  avatar_url: string;
  cover_url: string;
  created_at: string;
  updated_at: string;
}

/* ── Posts ─────────────────────────────────────────────────────── */

export interface Post {
  id: number;
  user_id: number;
  user: User;
  content: string;
  media: PostMedia[];
  like_count: number;
  comment_count: number;
  repost_count: number;
  bookmark_count: number;
  is_liked: boolean;
  is_reposted: boolean;
  is_bookmarked: boolean;
  created_at: string;
}

export interface PostMedia {
  id: number;
  media_url: string;
  media_type: string;
  order_index: number;
}

export interface Comment {
  id: number;
  user_id: number;
  user: User;
  post_id: number;
  parent_id: number | null;
  content: string;
  created_at: string;
}

/* ── Auth ──────────────────────────────────────────────────────── */

export interface AuthData {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
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
  follower_count: number;
  following_count: number;
}

/* ── API Response Wrapper ─────────────────────────────────────── */

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
