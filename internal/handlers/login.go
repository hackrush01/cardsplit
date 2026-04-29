package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Helper to render inline HTML errors for HTMX
func sendHTMXError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html")
	// Returning 200 OK so HTMX easily swaps it into the target div
	fmt.Fprintf(w, `<p class="text-red-500 text-sm font-medium">%s</p>`, msg)
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl := template.Must(template.ParseFiles("web/templates/login.html"))
			tmpl.Execute(w, nil)
			return
		}

		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == "" || password == "" {
				sendHTMXError(w, "Username and password are required")
				return
			}

			var hashedPassword sql.NullString
			err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hashedPassword)
			if err != nil {
				if err == sql.ErrNoRows {
					sendHTMXError(w, "User not found")
				} else {
					sendHTMXError(w, "Database error occurred")
				}
				return
			}

			if !hashedPassword.Valid || hashedPassword.String == "" {
				// FIRST TIME LOGIN: Hash and save the new password
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					sendHTMXError(w, "Error processing password")
					return
				}

				_, err = db.Exec("UPDATE users SET password_hash = ? WHERE username = ?", string(hash), username)
				if err != nil {
					sendHTMXError(w, "Error saving new password")
					return
				}
			} else {
				// SUBSEQUENT LOGIN: Verify the provided password against the stored hash
				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword.String), []byte(password))
				if err != nil {
					sendHTMXError(w, "Invalid password")
					return
				}
			}

			// LOGIN SUCCESS: Generate session
			sessionToken := generateSessionToken()

			_, err = db.Exec("INSERT INTO sessions (token, username, expires_at) VALUES (?, ?, ?)",
				sessionToken, username, time.Now().AddDate(0, 6, 0))
			if err != nil {
				sendHTMXError(w, "Error creating session")
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    sessionToken,
				Expires:  time.Now().AddDate(0, 6, 0),
				HttpOnly: true,
				Path:     "/",
				SameSite: http.SameSiteStrictMode, // Added for extra CSRF protection
			})

			// Tell HTMX to redirect the browser to the dashboard
			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func generateSessionToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
