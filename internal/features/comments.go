package features

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Comment struct {
	ID        int64
	PostID    int64
	AuthorID  int64
	Content   string
	CreatedAt time.Time
}

func CreateComment(ctx context.Context, db *sql.DB, postID, authorID int64, content string) (int64, error) {
	content = strings.TrimSpace(content)
	if postID <= 0 || authorID <= 0 || content == "" {
		return 0, errors.New("invalid comment data")
	}
	res, err := db.ExecContext(ctx, `
		INSERT INTO comments(post_id, author_id, content, created_at)
		VALUES(?,?,?,?)`, postID, authorID, content, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListCommentsByPostID(ctx context.Context, db *sql.DB, postID int64, limit, offset int) ([]Comment, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, post_id, author_id, content, created_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?`, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// DeleteComment deletes a comment (only by the author)
func DeleteComment(ctx context.Context, db *sql.DB, commentID, userID int64) error {
	if commentID <= 0 || userID <= 0 {
		return errors.New("invalid comment ID or user ID")
	}

	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if the user is the author of the comment
	var authorID int64
	err = tx.QueryRowContext(ctx, "SELECT author_id FROM comments WHERE id = ?", commentID).Scan(&authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("comment not found")
		}
		return err
	}

	if authorID != userID {
		return errors.New("you can only delete your own comments")
	}

	// Delete comment likes first
	_, err = tx.ExecContext(ctx, "DELETE FROM comment_likes WHERE comment_id = ?", commentID)
	if err != nil {
		return err
	}

	// Delete the comment
	_, err = tx.ExecContext(ctx, "DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
