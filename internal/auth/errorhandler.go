package auth

import (
	"html/template"
	"log"
	"net/http"
)

// HTTPErrorHandler handles HTTP errors with appropriate status codes and responses
type HTTPErrorHandler struct {
	templates *template.Template
	logger    *log.Logger
}

// NewHTTPErrorHandler creates a new HTTP error handler
func NewHTTPErrorHandler(templates *template.Template, logger *log.Logger) *HTTPErrorHandler {
	return &HTTPErrorHandler{
		templates: templates,
		logger:    logger,
	}
}

// Handle404 handles 404 Not Found errors
func (h *HTTPErrorHandler) Handle404(w http.ResponseWriter, r *http.Request) {
	h.handleError(w, r, http.StatusNotFound, "Page Not Found", "The page you are looking for does not exist.")
}

// Handle500 handles 500 Internal Server Error
func (h *HTTPErrorHandler) Handle500(w http.ResponseWriter, r *http.Request, err error) {
	if h.logger != nil {
		h.logger.Printf("Internal Server Error: %v | Path: %s | Method: %s", err, r.URL.Path, r.Method)
	}
	h.handleError(w, r, http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later.")
}

// Handle400 handles 400 Bad Request errors
func (h *HTTPErrorHandler) Handle400(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "The request could not be understood by the server."
	}
	h.handleError(w, r, http.StatusBadRequest, "Bad Request", message)
}

// handleError is the core error handling function
func (h *HTTPErrorHandler) handleError(w http.ResponseWriter, r *http.Request, statusCode int, title, message string) {
	w.WriteHeader(statusCode)

	// Try to render the error template
	if h.templates != nil {
		data := struct {
			Title      string
			StatusCode int
			StatusText string
			Message    string
			BackURL    string
			User       interface{} // Add User field for layout compatibility
		}{
			Title:      title,
			StatusCode: statusCode,
			StatusText: http.StatusText(statusCode),
			Message:    message,
			BackURL:    "/", // Default back URL
			User:       nil, // No user for error pages
		}

		// Set appropriate back URL based on the error
		if statusCode == http.StatusBadRequest {
			data.BackURL = "/"
		}

		if err := h.templates.ExecuteTemplate(w, "error.html", data); err != nil {
			// Fallback to plain text if template fails
			if h.logger != nil {
				h.logger.Printf("Failed to render error template: %v", err)
			}
			h.sendPlainTextError(w, statusCode, title, message)
		}
		return
	}

	// Fallback to plain text error
	h.sendPlainTextError(w, statusCode, title, message)
}

// sendPlainTextError sends a plain text error response as fallback
func (h *HTTPErrorHandler) sendPlainTextError(w http.ResponseWriter, statusCode int, title, message string) {
	w.Header().Set("Content-Type", "text/plain")
	// Note: WriteHeader already called in handleError, so don't call it again
	w.Write([]byte(title + "\n\n" + message))
}

// NotFoundHandler creates a handler function for 404 errors
func (h *HTTPErrorHandler) NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Handle404(w, r)
	}
}
