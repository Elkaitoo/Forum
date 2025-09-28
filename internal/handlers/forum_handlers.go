package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"forum/internal/auth"
	"forum/internal/features"
)

type ForumHandlers struct {
	db             *sql.DB
	authService    *auth.AuthService
	sessionService *auth.SessionService
	templates      *template.Template
	errorHandler   *auth.HTTPErrorHandler
}

func NewForumHandlers(db *sql.DB, authService *auth.AuthService, sessionService *auth.SessionService, templates *template.Template) *ForumHandlers {
	// Create error handler
	errorLogger := log.New(os.Stdout, "[FORUM-ERROR] ", log.LstdFlags|log.Lshortfile)
	errorHandler := auth.NewHTTPErrorHandler(templates, errorLogger)

	return &ForumHandlers{
		db:             db,
		authService:    authService,
		sessionService: sessionService,
		templates:      templates,
		errorHandler:   errorHandler,
	}
}

// Helper function to preserve query parameters and add anchor
func addAnchorToURL(baseURL, anchor string) string {
	if anchor == "" {
		return baseURL
	}
	
	// Parse the URL to handle existing query parameters
	u, err := url.Parse(baseURL)
	if err != nil {
		// If parsing fails, just append the anchor
		return baseURL + "#" + anchor
	}
	
	u.Fragment = anchor
	return u.String()
}

// HomeHandler displays the main forum page with posts
func (h *ForumHandlers) HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user if logged in
	var currentUser *auth.User
	var currentUserID int64
	if userID, ok := auth.GetUserFromContext(r); ok {
		user, err := h.authService.GetUserByID(userID)
		if err == nil {
			currentUser = user
			currentUserID = userID
		}
	}

	// Get query parameters for filtering
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("q")

	// Get posts with details from features layer
	posts, err := features.ListPostsWithDetails(r.Context(), h.db, features.ListOptions{
		CategoryName: category,
		Search:       search,
		Limit:        20,
		Offset:       0,
		OrderDesc:    true,
	}, currentUserID)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Get available categories for filters
	categories, err := features.GetAllCategories(r.Context(), h.db)
	if err != nil {
		categories = []features.Category{} // Empty if error
	}

	data := struct {
		Title      string
		User       *auth.User
		Posts      []features.PostWithDetails
		Categories []features.Category
		Filter     string
		Success    string
	}{
		Title:      "Forum",
		User:       currentUser,
		Posts:      posts,
		Categories: categories,
		Filter:     category,
	}

	// Check for success messages
	if r.URL.Query().Get("deleted") == "true" {
		data.Success = "Post deleted successfully."
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// CreatePostPageHandler shows the create post form
func (h *ForumHandlers) CreatePostPageHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	currentUser, err := h.authService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		// Get existing categories
		categories, err := features.GetAllCategories(r.Context(), h.db)
		if err != nil {
			categories = []features.Category{} // Empty if error
		}

		data := struct {
			Title              string
			User               *auth.User
			Error              string
			PostTitle          string
			PostContent        string
			Categories         string
			ExistingCategories []features.Category
		}{
			Title:              "Create Post",
			User:               currentUser,
			Error:              "",
			PostTitle:          "",
			PostContent:        "",
			Categories:         "",
			ExistingCategories: categories,
		}

		if err := h.templates.ExecuteTemplate(w, "create_post.html", data); err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))

		// Parse categories from multi-select dropdown
		selectedCategories := r.Form["existing_categories"]

		// Parse new categories from text input
		newCategoriesStr := strings.TrimSpace(r.FormValue("new_categories"))

		// Combine selected and new categories
		var categories []string

		// Add selected existing categories
		for _, cat := range selectedCategories {
			cat = strings.TrimSpace(cat)
			if cat != "" {
				categories = append(categories, cat)
			}
		}

		// Add new categories from text input
		if newCategoriesStr != "" {
			for _, cat := range strings.Split(newCategoriesStr, ",") {
				cat = strings.TrimSpace(cat)
				if cat != "" {
					categories = append(categories, cat)
				}
			}
		}

		// Create a categories string for template display
		categoriesStr := strings.Join(categories, ", ")

		// Validate input
		var errorMsg string
		if strings.TrimSpace(title) == "" {
			errorMsg = "Post title is required"
		} else if strings.TrimSpace(content) == "" {
			errorMsg = "Post content is required"
		}

		if errorMsg != "" {
			// Get existing categories for the error response
			existingCategories, _ := features.GetAllCategories(r.Context(), h.db)

			data := struct {
				Title              string
				User               *auth.User
				Error              string
				PostTitle          string
				PostContent        string
				Categories         string
				ExistingCategories []features.Category
			}{
				Title:              "Create Post",
				User:               currentUser,
				Error:              errorMsg,
				PostTitle:          title,
				PostContent:        content,
				Categories:         categoriesStr,
				ExistingCategories: existingCategories,
			}

			if err := h.templates.ExecuteTemplate(w, "create_post.html", data); err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
			return
		}

		// Create post
		postID, err := features.CreatePost(r.Context(), h.db, userID, title, content, categories)
		if err != nil {
			// Get existing categories for the error response
			existingCategories, _ := features.GetAllCategories(r.Context(), h.db)

			data := struct {
				Title              string
				User               *auth.User
				Error              string
				PostTitle          string
				PostContent        string
				Categories         string
				ExistingCategories []features.Category
			}{
				Title:              "Create Post",
				User:               currentUser,
				Error:              "Failed to create post: " + err.Error(),
				PostTitle:          title,
				PostContent:        content,
				Categories:         categoriesStr,
				ExistingCategories: existingCategories,
			}

			if err := h.templates.ExecuteTemplate(w, "create_post.html", data); err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
			return
		}

		// Redirect to the new post
		http.Redirect(w, r, "/post/"+strconv.FormatInt(postID, 10), http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// PostDetailHandler shows a single post with comments
func (h *ForumHandlers) PostDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path /post/123
	path := strings.TrimPrefix(r.URL.Path, "/post/")
	postID, err := strconv.ParseInt(path, 10, 64)
	if err != nil || postID <= 0 {
		h.errorHandler.Handle404(w, r)
		return
	}

	// Get current user if logged in
	var currentUser *auth.User
	var currentUserID int64
	if userID, ok := auth.GetUserFromContext(r); ok {
		user, err := h.authService.GetUserByID(userID)
		if err == nil {
			currentUser = user
			currentUserID = userID
		}
	}

	// Get post details
	post, err := features.GetPostWithDetails(r.Context(), h.db, postID, currentUserID)
	if err != nil {
		h.errorHandler.Handle404(w, r)
		return
	}

	// Get comments with details
	comments, err := features.ListCommentsWithDetails(r.Context(), h.db, postID, currentUserID)
	if err != nil {
		h.errorHandler.Handle500(w, r, err)
		return
	}

	data := struct {
		Title        string
		User         *auth.User
		Post         *features.PostWithDetails
		Comments     []features.CommentWithDetails
		Success      string
		CommentError string
	}{
		Title:        post.Title,
		User:         currentUser,
		Post:         post,
		Comments:     comments,
		Success:      r.URL.Query().Get("success"),
		CommentError: r.URL.Query().Get("comment_error"),
	}

	if err := h.templates.ExecuteTemplate(w, "post_detail.html", data); err != nil {
		h.errorHandler.Handle500(w, r, err)
		return
	}
}

// LikePostHandler handles post like/dislike actions
func (h *ForumHandlers) LikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	action := r.FormValue("action")
	var reaction int
	switch action {
	case "like":
		reaction = 1
	case "dislike":
		reaction = -1
	default:
		reaction = 0
	}

	if err := features.TogglePostReaction(r.Context(), h.db, userID, postID, reaction); err != nil {
		http.Error(w, "Failed to update reaction", http.StatusInternalServerError)
		return
	}

	// Get anchor for scroll position
	anchor := r.FormValue("anchor")

	// Redirect back to the post or previous page with anchor
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/post/" + strconv.FormatInt(postID, 10)
	}
	
	// Add anchor using helper function
	redirectURL := addAnchorToURL(referer, anchor)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// LikeCommentHandler handles comment like/dislike actions
func (h *ForumHandlers) LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.ParseInt(r.FormValue("comment_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	postID, _ := strconv.ParseInt(r.FormValue("post_id"), 10, 64)

	action := r.FormValue("action")
	var reaction int
	switch action {
	case "like":
		reaction = 1
	case "dislike":
		reaction = -1
	default:
		reaction = 0
	}

	if err := features.ToggleCommentReaction(r.Context(), h.db, userID, commentID, reaction); err != nil {
		http.Error(w, "Failed to update reaction", http.StatusInternalServerError)
		return
	}

	// Get anchor for scroll position
	anchor := r.FormValue("anchor")

	// Redirect back to the post with anchor
	redirectURL := "/post/" + strconv.FormatInt(postID, 10)
	redirectURL = addAnchorToURL(redirectURL, anchor)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// AddCommentHandler handles adding comments to posts
func (h *ForumHandlers) AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		// Redirect back with a custom error and anchor to comments section
		http.Redirect(w, r, "/post/"+strconv.FormatInt(postID, 10)+"?comment_error=Comment content is required#comments-section", http.StatusSeeOther)
		return
	}

	// Create the comment
	commentID, err := features.CreateComment(r.Context(), h.db, postID, userID, content)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post with anchor to the new comment
	redirectURL := "/post/" + strconv.FormatInt(postID, 10) + "#comment-" + strconv.FormatInt(commentID, 10)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// DeletePostHandler handles post deletion requests
func (h *ForumHandlers) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Delete the post
	err = features.DeletePost(r.Context(), h.db, postID, userID)
	if err != nil {
		http.Error(w, "Failed to delete post: "+err.Error(), http.StatusForbidden)
		return
	}

	// Redirect to home page with success message
	http.Redirect(w, r, "/?deleted=true", http.StatusSeeOther)
}

// DeleteCommentHandler handles comment deletion requests
func (h *ForumHandlers) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.ParseInt(r.FormValue("comment_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Delete the comment
	err = features.DeleteComment(r.Context(), h.db, commentID, userID)
	if err != nil {
		http.Error(w, "Failed to delete comment: "+err.Error(), http.StatusForbidden)
		return
	}

	// Redirect back to the post's comments section
	redirectURL := "/post/" + strconv.FormatInt(postID, 10) + "?comment_deleted=true#comments-section"
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
