package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/hackrush01/cardsplit/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type LoginTemplateData struct {
	Users []string
}

// Helper to render inline HTML errors for HTMX
func sendHTMXError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<p class="text-red-500 text-sm font-medium">%s</p>`, msg)
}

func RenderLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := storage.GetAllUsers(db)
		if err != nil {
			http.Error(w, "Load users", http.StatusInternalServerError)
			return
		}

		data := LoginTemplateData{
			Users: users,
		}

		tmpl, err := template.ParseFiles("web/templates/login.html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, data)
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			RenderLogin(db)(w, r)
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
				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword.String), []byte(password))
				if err != nil {
					sendHTMXError(w, "Invalid password")
					return
				}
			}

			st := generateSessionToken()
			_, err = db.Exec("INSERT INTO sessions (token, username, expires_at) VALUES (?, ?, ?)",
				st, username, time.Now().AddDate(0, 6, 0))
			if err != nil {
				sendHTMXError(w, "Error creating session")
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    st,
				Expires:  time.Now().AddDate(0, 6, 0),
				HttpOnly: true,
				Path:     "/",
				SameSite: http.SameSiteStrictMode, // Added for extra CSRF protection
			})

			if storage.GetUserRole(db, username) == "admin" {
				w.Header().Set("HX-Redirect", "/dashboard")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.Header().Set("HX-Redirect", "/statement")
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func generateSessionToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
