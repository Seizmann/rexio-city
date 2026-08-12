/**
 * Centralised constants for RexiO City frontend.
 * API paths, client routes, and validation limits from the PRD.
 */

/* ── API Endpoint Paths ───────────────────────────────────────── */
/* These are relative paths — Next.js middleware proxies them to the Go backend. */

export const API = {
  // Auth
  AUTH_SIGNUP: '/api/auth/signup',
  AUTH_LOGIN: '/api/auth/login',
  AUTH_REFRESH: '/api/auth/refresh',

  // Users
  USERS_ME: '/api/users/me',
  USER: (username: string) => `/api/users/${username}`,
  SEARCH: '/api/search',

  // Posts
  POSTS: '/api/posts',
  POST: (id: number | string) => `/api/posts/${id}`,
  POST_LIKE: (id: number | string) => `/api/posts/${id}/like`,
  POST_COMMENTS: (id: number | string) => `/api/posts/${id}/comments`,
  POST_REPOST: (id: number | string) => `/api/posts/${id}/repost`,
  POST_BOOKMARK: (id: number | string) => `/api/posts/${id}/bookmark`,

  // Feed
  FEED: '/api/feed',

  // Follow
  FOLLOW: (id: number | string) => `/api/users/${id}/follow`,
  FOLLOWERS: (id: number | string) => `/api/users/${id}/followers`,
  FOLLOWING: (id: number | string) => `/api/users/${id}/following`,
  FOLLOW_COUNTS: (id: number | string) => `/api/users/${id}/follow-counts`,
  IS_FOLLOWING: (id: number | string) => `/api/users/${id}/is-following`,

  // Media
  MEDIA_UPLOAD: '/api/media/upload',

  // Notifications
  NOTIFICATIONS: '/api/notifications',
  NOTIFICATION_READ: (id: number | string) => `/api/notifications/${id}/read`,
  NOTIFICATIONS_READ_ALL: '/api/notifications/read-all',
  NOTIFICATIONS_UNREAD: '/api/notifications/unread-count',
} as const;

/* ── Client Route Paths ───────────────────────────────────────── */

export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  SIGNUP: '/signup',
  PROFILE: (username: string) => `/${username.startsWith('@') ? username.slice(1) : username}`,
  POST: (id: number | string) => `/post/${id}`,
  MESSAGES: '/messages',
  NOTIFICATIONS: '/notifications',
} as const;

/* ── Validation Limits (from PRD §5) ──────────────────────────── */

export const LIMITS = {
  POST_MAX_CHARS: 500,
  BIO_MAX_CHARS: 160,
  DISPLAY_NAME_MAX_CHARS: 50,
  USERNAME_MIN_CHARS: 3,
  USERNAME_MAX_CHARS: 15,
  PASSWORD_MIN_CHARS: 8,
  COMMENT_MAX_CHARS: 500,
} as const;

/* ── Pagination ───────────────────────────────────────────────── */

export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
} as const;
