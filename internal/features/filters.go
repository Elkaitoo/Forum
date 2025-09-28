package features

import (
	"context"
	"database/sql"
)

// التفافات مريحة حول ListPosts لتوافق المطلوب

func ListPostsByCategory(ctx context.Context, db *sql.DB, category string, limit, offset int) ([]Post, error) {
	return ListPosts(ctx, db, ListOptions{
		CategoryName: category,
		Limit:        limit,
		Offset:       offset,
		OrderDesc:    true,
	})
}

func ListPostsByAuthor(ctx context.Context, db *sql.DB, authorID int64, limit, offset int) ([]Post, error) {
	return ListPosts(ctx, db, ListOptions{
		AuthorID:  authorID,
		Limit:     limit,
		Offset:    offset,
		OrderDesc: true,
	})
}

func ListPostsLikedByUser(ctx context.Context, db *sql.DB, userID int64, limit, offset int) ([]Post, error) {
	return ListPosts(ctx, db, ListOptions{
		LikedByUser: userID,
		Limit:       limit,
		Offset:      offset,
		OrderDesc:   true,
	})
}
