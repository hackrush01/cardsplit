package middleware

import (
	"context"
	"database/sql"
	"net/http"
)

// contextKey is used to prevent collisions in the context payload
type contextKey string

const UsernameKey contextKey = "username"

func AuthMiddleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract User Identity from Session
		cookie, err := r.Cookie("session_token")
		if err != nil {
			handleUnauthorized(w, r)
			return
		}

		// Query the sessions table to verify the token's validity and retrieve the username
		var username string
		err = db.QueryRow("SELECT username FROM sessions WHERE token = ?", cookie.Value).Scan(&username)
		if err != nil {
			// Token is invalid or expired
			handleUnauthorized(w, r)
			return
		}

		// Save this username directly into the Go http.Request context
		ctx := context.WithValue(r.Context(), UsernameKey, username)

		// Pass the new context to the downstream handlers
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// handleUnauthorized handles both standard and HTMX redirects securely
func handleUnauthorized(w http.ResponseWriter, r *http.Request) {
	// 4. Secure HTMX Interactions
	if r.Header.Get("HX-Request") == "true" {
		// Intercept the redirect and return a 204 No Content status alongside an HX-Redirect response header
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Standard redirect for non-HTMX requests
	http.Redirect(w, r, "/", http.StatusFound)
}
