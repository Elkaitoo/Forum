package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"forum/internal/auth"
	"forum/internal/features"
)

type FilterHandlers struct {
	db             *sql.DB
	sessionService *auth.SessionService
	templates      *template.Template
}

func NewFilterHandlers(db *sql.DB, sessionService *auth.SessionService, templates *template.Template) *FilterHandlers {
	return &FilterHandlers{
		db:             db,
		sessionService: sessionService,
		templates:      templates,
	}
}

// MyPostsHandler displays posts created by the current user
func (h *FilterHandlers) MyPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's posts
	posts, err := features.GetPostsByUserID(r.Context(), h.db, userID)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Get user info
	authService := auth.NewAuthService(h.db)
	currentUser, err := authService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
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
		Title:      "My Posts",
		User:       currentUser,
		Posts:      posts,
		Categories: categories,
		Filter:     "my-posts",
		Success:    "",
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// LikedPostsHandler displays posts liked by the current user
func (h *FilterHandlers) LikedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get posts liked by user
	posts, err := features.GetLikedPostsByUserID(r.Context(), h.db, userID)
	if err != nil {
		http.Error(w, "Failed to load liked posts", http.StatusInternalServerError)
		return
	}

	// Get user info
	authService := auth.NewAuthService(h.db)
	currentUser, err := authService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
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
		Title:      "Liked Posts",
		User:       currentUser,
		Posts:      posts,
		Categories: categories,
		Filter:     "liked-posts",
		Success:    "",
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
