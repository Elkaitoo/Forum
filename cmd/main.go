package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/handlers"
)

func main() {
	// Initialize database
	db, err := database.NewDB(database.DefaultConfig())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize database tables
	if err := db.InitializeDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Create template functions for better date formatting
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006 at 3:04 PM")
		},
		"timeAgo": func(t time.Time) string {
			duration := time.Since(t)

			if duration.Hours() < 1 {
				if duration.Minutes() < 1 {
					return "just now"
				}
				return fmt.Sprintf("%.0f minutes ago", duration.Minutes())
			} else if duration.Hours() < 24 {
				return fmt.Sprintf("%.0f hours ago", duration.Hours())
			} else if duration.Hours() < 168 { // 7 days
				return fmt.Sprintf("%.0f days ago", duration.Hours()/24)
			} else {
				return t.Format("Jan 2, 2006")
			}
		},
	}

	// Load HTML templates with custom functions
	templates := template.New("").Funcs(funcMap)
	templates, err = templates.ParseFiles(
		"web/templates/layout.html",
		"web/templates/index.html",
		"web/templates/login.html",
		"web/templates/register.html",
		"web/templates/create_post.html",
		"web/templates/post_detail.html",
		"web/templates/error.html",
	)
	if err != nil {
		log.Fatal("Failed to load templates:", err)
	}

	// Initialize services
	authService := auth.NewAuthService(db.DB)
	sessionService := auth.NewSessionService(db.DB)
	authMiddleware := auth.NewMiddleware(sessionService)

	// Initialize error handler
	errorLogger := log.New(log.Writer(), "[ERROR] ", log.LstdFlags|log.Lshortfile)
	errorHandler := auth.NewHTTPErrorHandler(templates, errorLogger)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, sessionService, templates)
	forumHandlers := handlers.NewForumHandlers(db.DB, authService, sessionService, templates)
	filterHandlers := handlers.NewFilterHandlers(db.DB, sessionService, templates)

	// Create a custom mux to handle 404 errors
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/", authMiddleware.OptionalAuth(forumHandlers.HomeHandler))
	mux.HandleFunc("/login", authHandlers.LoginHandler)
	mux.HandleFunc("/register", authHandlers.RegisterHandler)
	mux.HandleFunc("/logout", authHandlers.LogoutHandler)

	// Protected routes
	mux.HandleFunc("/create-post", authMiddleware.RequireAuth(forumHandlers.CreatePostPageHandler))
	mux.HandleFunc("/add-comment", authMiddleware.RequireAuth(forumHandlers.AddCommentHandler))
	mux.HandleFunc("/delete-post", authMiddleware.RequireAuth(forumHandlers.DeletePostHandler))
	mux.HandleFunc("/delete-comment", authMiddleware.RequireAuth(forumHandlers.DeleteCommentHandler))
	mux.HandleFunc("/my-posts", authMiddleware.RequireAuth(filterHandlers.MyPostsHandler))
	mux.HandleFunc("/liked-posts", authMiddleware.RequireAuth(filterHandlers.LikedPostsHandler))
	mux.HandleFunc("/post/", authMiddleware.OptionalAuth(forumHandlers.PostDetailHandler))
	mux.HandleFunc("/like-post", authMiddleware.RequireAuth(forumHandlers.LikePostHandler))
	mux.HandleFunc("/like-comment", authMiddleware.RequireAuth(forumHandlers.LikeCommentHandler))

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Create a wrapper that handles 404 errors
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path matches any registered routes
		if r.URL.Path != "/" && !routeExists(r.URL.Path) && !isStaticFile(r.URL.Path) && !isPostDetail(r.URL.Path) {
			errorHandler.Handle404(w, r)
			return
		}

		// Serve the request normally
		mux.ServeHTTP(w, r)
	})

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// routeExists checks if a route is registered
func routeExists(path string) bool {
	validRoutes := []string{
		"/", "/login", "/register", "/logout",
		"/create-post", "/add-comment", "/delete-post", "/delete-comment",
		"/my-posts", "/liked-posts", "/like-post", "/like-comment",
	}

	for _, route := range validRoutes {
		if path == route {
			return true
		}
	}
	return false
}

// isStaticFile checks if the path is for a static file
func isStaticFile(path string) bool {
	return len(path) > 8 && path[:8] == "/static/"
}

// isPostDetail checks if the path is for a post detail page
func isPostDetail(path string) bool {
	return len(path) > 6 && path[:6] == "/post/"
}
