package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"forum/internal/auth"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	authService    *auth.AuthService
	sessionService *auth.SessionService
	templates      *template.Template
	errorHandler   *auth.HTTPErrorHandler
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService *auth.AuthService, sessionService *auth.SessionService, templates *template.Template) *AuthHandlers {
	// Create error handler
	errorLogger := log.New(os.Stdout, "[AUTH-ERROR] ", log.LstdFlags|log.Lshortfile)
	errorHandler := auth.NewHTTPErrorHandler(templates, errorLogger)

	return &AuthHandlers{
		authService:    authService,
		sessionService: sessionService,
		templates:      templates,
		errorHandler:   errorHandler,
	}
}

// RegisterHandler handles both GET (show form) and POST (process registration)
func (h *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Show registration form
		data := struct {
			Title    string
			Error    string
			Email    string
			Username string
		}{
			Title:    "Register",
			Error:    "",
			Email:    "",
			Username: "",
		}

		if err := h.templates.ExecuteTemplate(w, "register.html", data); err != nil {
			h.errorHandler.Handle500(w, r, err)
			return
		}

	case http.MethodPost:
		// Process registration
		if err := r.ParseForm(); err != nil {
			h.errorHandler.Handle400(w, r, "Invalid form data")
			return
		}

		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Validate required fields with custom messages
		var errorMsg string
		if strings.TrimSpace(email) == "" {
			errorMsg = "Email is required"
		} else if strings.TrimSpace(username) == "" {
			errorMsg = "Username is required"
		} else if strings.TrimSpace(password) == "" {
			errorMsg = "Password is required"
		}

		if errorMsg != "" {
			data := struct {
				Title    string
				Error    string
				Email    string
				Username string
			}{
				Title:    "Register",
				Error:    errorMsg,
				Email:    email,
				Username: username,
			}

			if err := h.templates.ExecuteTemplate(w, "register.html", data); err != nil {
				h.errorHandler.Handle500(w, r, err)
			}
			return
		}

		// Attempt to register user
		err := h.authService.RegisterUser(email, username, password)
		if err != nil {
			data := struct {
				Title    string
				Error    string
				Email    string
				Username string
			}{
				Title:    "Register",
				Error:    err.Error(),
				Email:    email,
				Username: username,
			}

			if err := h.templates.ExecuteTemplate(w, "register.html", data); err != nil {
				h.errorHandler.Handle500(w, r, err)
			}
			return
		}

		// Registration successful, redirect to login
		http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LoginHandler handles both GET (show form) and POST (process login)
func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Show login form
		data := struct {
			Title      string
			Error      string
			Success    string
			Email      string
			Registered bool
		}{
			Title:      "Login",
			Email:      "",
			Registered: r.URL.Query().Get("registered") == "true",
		}

		if data.Registered {
			data.Success = "Registration successful! Please log in."
		}

		// Check for logout success message
		if r.URL.Query().Get("logout") == "success" {
			data.Success = "You have been successfully logged out."
		}

		if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
			h.errorHandler.Handle500(w, r, err)
			return
		}

	case http.MethodPost:
		// Process login
		if err := r.ParseForm(); err != nil {
			h.errorHandler.Handle400(w, r, "Invalid form data")
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Validate required fields with custom messages
		var errorMsg string
		if strings.TrimSpace(email) == "" {
			errorMsg = "Email is required"
		} else if strings.TrimSpace(password) == "" {
			errorMsg = "Password is required"
		}

		if errorMsg != "" {
			data := struct {
				Title   string
				Error   string
				Success string
				Email   string
			}{
				Title:   "Login",
				Error:   errorMsg,
				Success: "",
				Email:   email,
			}

			if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
				h.errorHandler.Handle500(w, r, err)
			}
			return
		}

		// Authenticate user
		userID, err := h.authService.AuthenticateUser(email, password)
		if err != nil {
			data := struct {
				Title   string
				Error   string
				Success string
				Email   string
			}{
				Title:   "Login",
				Error:   err.Error(),
				Success: "", // Empty success message for error case
				Email:   email,
			}

			if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
				h.errorHandler.Handle500(w, r, err)
			}
			return
		}

		// Create session
		sessionToken, err := h.sessionService.CreateSession(userID)
		if err != nil {
			h.errorHandler.Handle500(w, r, err)
			return
		}

		// Set session cookie
		h.sessionService.SetSessionCookie(w, sessionToken)

		// Redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LogoutHandler handles logout requests
func (h *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete session from database
		h.sessionService.DeleteSession(cookie.Value)
	}

	// Clear session cookie
	h.sessionService.ClearSessionCookie(w)

	// Redirect to login page with success message
	http.Redirect(w, r, "/login?logout=success", http.StatusSeeOther)
}
