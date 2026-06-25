package handlers

import (
	"database/sql"
	"net/http"

	"github.com/hackrush01/cardsplit/internal/storage"
)

func AdminDashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		users, err := storage.GetAllUserStatuses(db)
		if err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}

		data := struct {
			Users []storage.UserStatus
		}{
			Users: users,
		}

		RenderTemplate(w, data, "web/templates/admin_dashboard.html")
	}
}
