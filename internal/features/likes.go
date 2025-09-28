package features

import (
	"context"
	"database/sql"
	"errors"
)

// reaction: 1 like, -1 dislike, 0 remove
func TogglePostReaction(ctx context.Context, db *sql.DB, userID, postID int64, reaction int) error {
	if userID <= 0 || postID <= 0 {
		return errors.New("invalid ids")
	}
	if reaction != -1 && reaction != 0 && reaction != 1 {
		return errors.New("invalid reaction")
	}
	// استخدم UPSERT لـ SQLite (ON CONFLICT)
	if reaction == 0 {
		_, err := db.ExecContext(ctx, `DELETE FROM post_likes WHERE user_id=? AND post_id=?`, userID, postID)
		return err
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO post_likes(user_id, post_id, reaction)
		VALUES(?,?,?)
		ON CONFLICT(user_id, post_id) DO UPDATE SET reaction=excluded.reaction
	`, userID, postID, reaction)
	return err
}

func ToggleCommentReaction(ctx context.Context, db *sql.DB, userID, commentID int64, reaction int) error {
	if userID <= 0 || commentID <= 0 {
		return errors.New("invalid ids")
	}
	if reaction != -1 && reaction != 0 && reaction != 1 {
		return errors.New("invalid reaction")
	}
	if reaction == 0 {
		_, err := db.ExecContext(ctx, `DELETE FROM comment_likes WHERE user_id=? AND comment_id=?`, userID, commentID)
		return err
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO comment_likes(user_id, comment_id, reaction)
		VALUES(?,?,?)
		ON CONFLICT(user_id, comment_id) DO UPDATE SET reaction=excluded.reaction
	`, userID, commentID, reaction)
	return err
}

type Reactions struct {
	Likes    int
	Dislikes int
}

func CountPostReactions(ctx context.Context, db *sql.DB, postID int64) (Reactions, error) {
	var r Reactions
	row := db.QueryRowContext(ctx, `
		SELECT
			SUM(CASE WHEN reaction=1 THEN 1 ELSE 0 END) AS likes,
			SUM(CASE WHEN reaction=-1 THEN 1 ELSE 0 END) AS dislikes
		FROM post_likes WHERE post_id = ?`, postID)
	if err := row.Scan(&r.Likes, &r.Dislikes); err != nil {
		return Reactions{}, err
	}
	return r, nil
}

func CountCommentReactions(ctx context.Context, db *sql.DB, commentID int64) (Reactions, error) {
	var r Reactions
	row := db.QueryRowContext(ctx, `
		SELECT
			SUM(CASE WHEN reaction=1 THEN 1 ELSE 0 END) AS likes,
			SUM(CASE WHEN reaction=-1 THEN 1 ELSE 0 END) AS dislikes
		FROM comment_likes WHERE comment_id = ?`, commentID)
	if err := row.Scan(&r.Likes, &r.Dislikes); err != nil {
		return Reactions{}, err
	}
	return r, nil
}
