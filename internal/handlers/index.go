package handlers

import (
	"net/http"
	"text/template"
)

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/admin_dashboard.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
