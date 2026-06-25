package api

import (
	"database/sql"
	"net/http"

	"github.com/hackrush01/cardsplit/internal/storage"
)

// ResetPasswordHandler handles HTMX requests to reset a user's password.
func ResetPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username := r.FormValue("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		err := storage.ResetUserPassword(db, username)
		if err != nil {
			http.Error(w, "Failed to reset password", http.StatusInternalServerError)
			return
		}

		// Refresh the page to reflect the new state.
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
	}
}
