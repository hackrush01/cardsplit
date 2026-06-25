package handlers

import (
	"database/sql"
	"net/http"
	"text/template"

	"github.com/hackrush01/cardsplit/internal/storage"
)

func AdminDashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/admin_dashboard.html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

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

		tmpl.Execute(w, data)
	}
}
