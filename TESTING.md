# TESTING.md — Testing Procedures for RexiO City

Guide for manual QA and automated testing.

---

## Manual QA Checklist

### Authentication Flow

- [ ] Sign up with email/password
- [ ] Verify email via Brevo link
- [ ] Login with email/password
- [ ] Social login (Google)
- [ ] Social login (GitHub)
- [ ] Refresh token rotation works
- [ ] Logout clears session

### Posts

- [ ] Create text post (max 500 chars)
- [ ] Create post with photo
- [ ] Create post with video
- [ ] Create post with voice
- [ ] Create post with link (preview appears)
- [ ] Delete own post
- [ ] Cannot delete others' posts
- [ ] Post appears in feed after creation

### Feed

- [ ] Following tab shows only followed users' posts
- [ ] For You tab shows algorithmic content
- [ ] Pagination works (load more)
- [ ] Empty state shows helpful message

### Engagement

- [ ] Like/unlike toggle works
- [ ] Comment on post
- [ ] Reply to comment (nested)
- [ ] Repost with/without comment
- [ ] Bookmark/unbookmark
- [ ] Counts update correctly

### Profiles

- [ ] View own profile
- [ ] View other users' profiles
- [ ] Edit profile (bio, avatar, cover)
- [ ] Follow/unfollow
- [ ] Following/followers counts update
- [ ] Profile tabs (Posts, Replies, Media)

### DMs

- [ ] Start new conversation
- [ ] Send message
- [ ] Receive message (real-time)
- [ ] Typing indicator appears
- [ ] Read receipt updates
- [ ] Message history loads
- [ ] Delete conversation

### Notifications

- [ ] New follower notification
- [ ] Like notification
- [ ] Comment notification
- [ ] DM reply notification
- [ ] Mark as read

### Admin

- [ ] Login to admin panel
- [ ] View user list
- [ ] Ban/unban user
- [ ] Delete post
- [ ] View system settings

---

## Automated Tests

### Go Backend

Run with: `go test ./...`

Critical paths requiring tests:
- Auth (login, signup, token validation)
- Feed scoring algorithm
- Follow system
- DM message encryption/decryption

### Frontend

No mandatory E2E in V1. Manually verify flows above.

---

## CI Checks

GitHub Actions runs on every push to `dev`:

1. `gofmt` check
2. `go vet` check
3. `go test ./...`
4. `eslint` for frontend and admin
5. `tsc --noEmit` for TypeScript checks
6. gitleaks scan (prevent secret commits)

---

*Last updated: 2026-08-12*
