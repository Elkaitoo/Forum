package database

import (
	"time"
)

// User represents a forum user
type User struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

// Session represents a user session for authentication
type Session struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// Post represents a forum post
type Post struct {
	ID        int64     `db:"id"`
	AuthorID  int64     `db:"author_id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

// Category represents a post category
type Category struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// PostCategory represents the many-to-many relationship between posts and categories
type PostCategory struct {
	PostID     int64 `db:"post_id"`
	CategoryID int64 `db:"category_id"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        int64     `db:"id"`
	PostID    int64     `db:"post_id"`
	AuthorID  int64     `db:"author_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

// PostLike represents a like/dislike on a post
type PostLike struct {
	UserID   int64 `db:"user_id"`
	PostID   int64 `db:"post_id"`
	Reaction int   `db:"reaction"` // 1 for like, -1 for dislike
}

// CommentLike represents a like/dislike on a comment
type CommentLike struct {
	UserID    int64 `db:"user_id"`
	CommentID int64 `db:"comment_id"`
	Reaction  int   `db:"reaction"` // 1 for like, -1 for dislike
}
