package models

import "time"

// User represents a platform user
type User struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	Username      string     `json:"username" gorm:"uniqueIndex;size:15"`
	DisplayName   *string    `json:"display_name" gorm:"size:50"`
	Bio           *string    `json:"bio" gorm:"size:160"`
	AvatarURL     *string    `json:"avatar_url"`
	CoverURL      *string    `json:"cover_url"`
	Email         *string    `json:"email" gorm:"uniqueIndex"`
	PasswordHash  string     `json:"-" gorm:"size:255"`
	EmailVerified bool       `json:"email_verified" gorm:"default:false"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Post represents a social post with a 16-character random public_id
type Post struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	PublicID  string     `json:"public_id" gorm:"uniqueIndex;size:32"`
	UserID    uint       `json:"user_id"`
	Content   string     `json:"content" gorm:"size:500"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
	User      User       `json:"user" gorm:"foreignKey:UserID"`
}

// PostMedia represents media attached to a post
type PostMedia struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PostID    uint      `json:"post_id"`
	MediaURL  string    `json:"media_url"`
	MediaType string    `json:"media_type"` // photo, video, voice
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"created_at"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	PostID    uint      `json:"post_id"`
	ParentID  *uint     `json:"parent_id"` // for nested replies
	Content   string    `json:"content" gorm:"size:500"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

// Like represents a like on a post
type Like struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	PostID    uint      `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Repost represents a repost
type Repost struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	PostID    uint      `json:"post_id"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// Bookmark represents a saved post
type Bookmark struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	PostID    uint      `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Follow represents a follow relationship
type Follow struct {
	FollowerID uint      `json:"follower_id" gorm:"primaryKey"`
	FolloweeID uint      `json:"followee_id" gorm:"primaryKey"`
	CreatedAt  time.Time `json:"created_at"`
}

// DMConversation represents a DM thread
type DMConversation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

// DMParticipant represents a participant in a DM conversation
type DMParticipant struct {
	ConversationID uint `json:"conversation_id" gorm:"primaryKey"`
	UserID         uint `json:"user_id" gorm:"primaryKey"`
}

// DMMessage represents an encrypted DM message
type DMMessage struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ConversationID uint      `json:"conversation_id"`
	SenderID       uint      `json:"sender_id"`
	EncryptedData  []byte    `json:"-"` // AES-256-GCM ciphertext
	IV             []byte    `json:"iv"` // nonce
	CreatedAt      time.Time `json:"created_at"`
}

// Notification represents a user notification
type Notification struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id"`
	Type      string     `json:"type"` // follower, like, comment, repost, mention, dm_reply
	ActorID   *uint      `json:"actor_id"`
	PostID    *uint      `json:"post_id"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// Setting represents a DB-driven configuration value
type Setting struct {
	Key       string    `json:"key" gorm:"primaryKey;size:100"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token" gorm:"uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
