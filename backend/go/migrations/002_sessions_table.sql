-- Migration 002: Secure session tracking with rotation lineage
--
-- Replaces the unused refresh_tokens table with a full sessions table.
-- Key design decisions:
--   - token_hash: SHA-256 of the raw refresh token (never store raw tokens)
--   - parent_session_id: rotation lineage — each refresh creates a child session
--   - revoked_at: nullable; non-null means this session was revoked
--   - If a token_hash is presented and revoked_at IS NOT NULL, that's reuse
--     detection — revoke the entire session family (same user, same root lineage)

CREATE TABLE IF NOT EXISTS sessions (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- SHA-256 hex of the raw refresh token JWT; never store the raw token
    token_hash       VARCHAR(64) UNIQUE NOT NULL,
    -- Rotation lineage: points to the session this was rotated from
    parent_session_id BIGINT REFERENCES sessions(id) ON DELETE SET NULL,
    -- Device / network context at session creation time
    device_info      TEXT,       -- User-Agent header value
    ip_address       VARCHAR(45), -- IPv4 or IPv6
    -- Lifecycle timestamps
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    -- NULL = active; non-NULL = revoked (either by logout, rotation, or reuse detection)
    revoked_at       TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_parent     ON sessions(parent_session_id);

-- The old refresh_tokens table is superseded; drop it if it exists
DROP TABLE IF EXISTS refresh_tokens;
