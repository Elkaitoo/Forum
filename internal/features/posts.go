package features

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Post struct {
	ID         int64
	AuthorID   int64
	Title      string
	Content    string
	CreatedAt  time.Time
	Categories []string // أسماء التصنيفات المرتبطة (اختياري للعرض)
}

func CreatePost(ctx context.Context, db *sql.DB, authorID int64, title, content string, categoryNames []string) (int64, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if authorID <= 0 || title == "" || content == "" {
		return 0, errors.New("invalid post data")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO posts(author_id, title, content, created_at) VALUES(?,?,?,?)`,
		authorID, title, content, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if len(categoryNames) > 0 {
		if err := associateCategoriesTx(ctx, tx, postID, categoryNames); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return postID, nil
}

func GetPostByID(ctx context.Context, db *sql.DB, id int64) (*Post, error) {
	row := db.QueryRowContext(ctx, `
		SELECT p.id, p.author_id, p.title, p.content, p.created_at
		FROM posts p WHERE p.id = ?`, id)
	var p Post
	if err := row.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.CreatedAt); err != nil {
		return nil, err
	}

	// جلب التصنيفات
	cats, err := listCategoriesForPost(ctx, db, p.ID)
	if err == nil {
		p.Categories = cats
	}
	return &p, nil
}

type ListOptions struct {
	Limit        int
	Offset       int
	CategoryName string // فلترة حسب تصنيف واحد (إن وُجد)
	AuthorID     int64  // فلترة “منشوراتي”
	LikedByUser  int64  // فلترة “التي أعجبتني”
	Search       string // بحث بسيط بالعنوان/المحتوى (اختياري)
	OrderDesc    bool   // الأحدث أولاً
}

// ListPosts يدعم الفلاتر المطلوبة: التصنيفات / منشوراتي / المعجب بها
func ListPosts(ctx context.Context, db *sql.DB, opt ListOptions) ([]Post, error) {
	var args []any
	var sb strings.Builder

	sb.WriteString(`
	SELECT DISTINCT p.id, p.author_id, p.title, p.content, p.created_at
	FROM posts p
	LEFT JOIN post_categories pc ON pc.post_id = p.id
	LEFT JOIN categories c ON c.id = pc.category_id
	`)

	// لو فلترة "المعجب بها"
	if opt.LikedByUser > 0 {
		sb.WriteString(" JOIN post_likes pl ON pl.post_id = p.id AND pl.user_id = ? AND pl.reaction = 1 ")
		args = append(args, opt.LikedByUser)
	}

	var where []string
	if opt.CategoryName != "" {
		where = append(where, "c.name = ?")
		args = append(args, opt.CategoryName)
	}
	if opt.AuthorID > 0 {
		where = append(where, "p.author_id = ?")
		args = append(args, opt.AuthorID)
	}
	if s := strings.TrimSpace(opt.Search); s != "" {
		where = append(where, "(p.title LIKE ? OR p.content LIKE ?)")
		args = append(args, "%"+s+"%", "%"+s+"%")
	}

	if len(where) > 0 {
		sb.WriteString(" WHERE " + strings.Join(where, " AND "))
	}

	if opt.OrderDesc {
		sb.WriteString(" ORDER BY p.created_at DESC ")
	} else {
		sb.WriteString(" ORDER BY p.created_at ASC ")
	}

	limit := opt.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opt.Offset
	if offset < 0 {
		offset = 0
	}
	sb.WriteString(" LIMIT ? OFFSET ? ")
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		if cats, err := listCategoriesForPost(ctx, db, p.ID); err == nil {
			p.Categories = cats
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func associateCategoriesTx(ctx context.Context, tx *sql.Tx, postID int64, categoryNames []string) error {
	for _, raw := range categoryNames {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		var catID int64
		// تأكد وجود التصنيف، وإلا أنشئه
		err := tx.QueryRowContext(ctx, `SELECT id FROM categories WHERE name = ?`, name).Scan(&catID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res, err2 := tx.ExecContext(ctx, `INSERT INTO categories(name) VALUES(?)`, name)
				if err2 != nil {
					return err2
				}
				catID, err = res.LastInsertId()
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO post_categories(post_id, category_id) VALUES(?,?)`, postID, catID); err != nil {
			return err
		}
	}
	return nil
}

func listCategoriesForPost(ctx context.Context, db *sql.DB, postID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT c.name
		FROM post_categories pc
		JOIN categories c ON c.id = pc.category_id
		WHERE pc.post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// PostWithDetails extends Post with additional details for display
type PostWithDetails struct {
	Post
	Username      string
	LikesCount    int
	DislikesCount int
	CommentsCount int
	UserLiked     bool
	UserDisliked  bool
}

// CommentWithDetails extends Comment with additional details for display
type CommentWithDetails struct {
	Comment
	Username      string
	LikesCount    int
	DislikesCount int
	UserLiked     bool
	UserDisliked  bool
}

// Category represents a forum category
type Category struct {
	ID   int64
	Name string
}

// GetAllCategories returns all available categories
func GetAllCategories(ctx context.Context, db *sql.DB) ([]Category, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

// GetPostsByUserID returns posts created by a specific user
func GetPostsByUserID(ctx context.Context, db *sql.DB, userID int64) ([]PostWithDetails, error) {
	return ListPostsWithDetails(ctx, db, ListOptions{
		AuthorID:  userID,
		Limit:     100,
		OrderDesc: true,
	}, 0)
}

// GetLikedPostsByUserID returns posts liked by a specific user
func GetLikedPostsByUserID(ctx context.Context, db *sql.DB, userID int64) ([]PostWithDetails, error) {
	return ListPostsWithDetails(ctx, db, ListOptions{
		LikedByUser: userID,
		Limit:       100,
		OrderDesc:   true,
	}, userID)
}

// ListPostsWithDetails returns posts with additional details for display
func ListPostsWithDetails(ctx context.Context, db *sql.DB, opt ListOptions, currentUserID int64) ([]PostWithDetails, error) {
	posts, err := ListPosts(ctx, db, opt)
	if err != nil {
		return nil, err
	}

	var result []PostWithDetails
	for _, post := range posts {
		detail := PostWithDetails{Post: post}

		// Get username
		err := db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", post.AuthorID).Scan(&detail.Username)
		if err != nil {
			detail.Username = "Unknown"
		}

		// Get reaction counts
		reactions, err := CountPostReactions(ctx, db, post.ID)
		if err == nil {
			detail.LikesCount = reactions.Likes
			detail.DislikesCount = reactions.Dislikes
		}

		// Get comments count
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE post_id = ?", post.ID).Scan(&detail.CommentsCount)
		if err != nil {
			detail.CommentsCount = 0
		}

		// Check if current user liked/disliked this post
		if currentUserID > 0 {
			var reaction sql.NullInt64
			err = db.QueryRowContext(ctx, "SELECT reaction FROM post_likes WHERE user_id = ? AND post_id = ?", currentUserID, post.ID).Scan(&reaction)
			if err == nil && reaction.Valid {
				if reaction.Int64 == 1 {
					detail.UserLiked = true
				} else if reaction.Int64 == -1 {
					detail.UserDisliked = true
				}
			}
		}

		result = append(result, detail)
	}

	return result, nil
}

// ListCommentsWithDetails returns comments with additional details for display
func ListCommentsWithDetails(ctx context.Context, db *sql.DB, postID int64, currentUserID int64) ([]CommentWithDetails, error) {
	comments, err := ListCommentsByPostID(ctx, db, postID, 100, 0)
	if err != nil {
		return nil, err
	}

	var result []CommentWithDetails
	for _, comment := range comments {
		detail := CommentWithDetails{Comment: comment}

		// Get username
		err := db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", comment.AuthorID).Scan(&detail.Username)
		if err != nil {
			detail.Username = "Unknown"
		}

		// Get reaction counts
		reactions, err := CountCommentReactions(ctx, db, comment.ID)
		if err == nil {
			detail.LikesCount = reactions.Likes
			detail.DislikesCount = reactions.Dislikes
		}

		// Check if current user liked/disliked this comment
		if currentUserID > 0 {
			var reaction sql.NullInt64
			err = db.QueryRowContext(ctx, "SELECT reaction FROM comment_likes WHERE user_id = ? AND comment_id = ?", currentUserID, comment.ID).Scan(&reaction)
			if err == nil && reaction.Valid {
				if reaction.Int64 == 1 {
					detail.UserLiked = true
				} else if reaction.Int64 == -1 {
					detail.UserDisliked = true
				}
			}
		}

		result = append(result, detail)
	}

	return result, nil
}

// GetPostWithDetails returns a single post with all details
func GetPostWithDetails(ctx context.Context, db *sql.DB, postID int64, currentUserID int64) (*PostWithDetails, error) {
	post, err := GetPostByID(ctx, db, postID)
	if err != nil {
		return nil, err
	}

	detail := PostWithDetails{Post: *post}

	// Get username
	err = db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", post.AuthorID).Scan(&detail.Username)
	if err != nil {
		detail.Username = "Unknown"
	}

	// Get reaction counts
	reactions, err := CountPostReactions(ctx, db, post.ID)
	if err == nil {
		detail.LikesCount = reactions.Likes
		detail.DislikesCount = reactions.Dislikes
	}

	// Get comments count
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE post_id = ?", post.ID).Scan(&detail.CommentsCount)
	if err != nil {
		detail.CommentsCount = 0
	}

	// Check if current user liked/disliked this post
	if currentUserID > 0 {
		var reaction sql.NullInt64
		err = db.QueryRowContext(ctx, "SELECT reaction FROM post_likes WHERE user_id = ? AND post_id = ?", currentUserID, post.ID).Scan(&reaction)
		if err == nil && reaction.Valid {
			if reaction.Int64 == 1 {
				detail.UserLiked = true
			} else if reaction.Int64 == -1 {
				detail.UserDisliked = true
			}
		}
	}

	return &detail, nil
}

// DeletePost deletes a post and all its associated data (only by the author)
func DeletePost(ctx context.Context, db *sql.DB, postID, userID int64) error {
	if postID <= 0 || userID <= 0 {
		return errors.New("invalid post ID or user ID")
	}

	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if the user is the author of the post
	var authorID int64
	err = tx.QueryRowContext(ctx, "SELECT author_id FROM posts WHERE id = ?", postID).Scan(&authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("post not found")
		}
		return err
	}

	if authorID != userID {
		return errors.New("you can only delete your own posts")
	}

	// Delete in order: likes, comments, post_categories, then post
	// Delete post likes
	_, err = tx.ExecContext(ctx, "DELETE FROM post_likes WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	// Delete comment likes first
	_, err = tx.ExecContext(ctx, `
		DELETE FROM comment_likes 
		WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)
	`, postID)
	if err != nil {
		return err
	}

	// Delete comments
	_, err = tx.ExecContext(ctx, "DELETE FROM comments WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	// Delete post categories
	_, err = tx.ExecContext(ctx, "DELETE FROM post_categories WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	// Finally delete the post
	_, err = tx.ExecContext(ctx, "DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
