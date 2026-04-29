package api

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/hackrush01/cardsplit/internal/config"
	"github.com/hackrush01/cardsplit/internal/models"
	"github.com/hackrush01/cardsplit/internal/parsers"
	"github.com/hackrush01/cardsplit/internal/storage"
)

// UploadHandler processes the HTMX form submission
func UploadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Load the mapping config into memory
		mappingPath := os.Getenv("CONFIG_PATH")
		if mappingPath == "" {
			mappingPath = "./configs/card_mapping.json"
		}

		cardMapping, err := config.LoadCardMapping(mappingPath)
		if err != nil {
			http.Error(w, "Failed to load user configurations", http.StatusInternalServerError)
			return
		}

		// 2. Parse the uploaded file (or hardcoded file)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("statement")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		statement, err := parsers.ParseInfiniaCSV(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div class="rounded border border-red-300 bg-red-50 p-4 text-red-700"><strong>Error:</strong> %s</div>`, template.HTMLEscapeString(err.Error()))
			return
		}

		// 3. Map the Raw Bank Labels to the login username for this statement.
		for i, txn := range statement.Transactions {
			username, _, err := cardMapping.GetUserDetails("Infinia", txn.RawLabel)
			if err == nil {
				statement.Transactions[i].Username = username
			} else {
				statement.Transactions[i].Username = ""
			}
		}

		// 4. Save the parsed data to the database
		if err := storage.SaveStatement(db, statement); err != nil {
			http.Error(w, "Failed to save statement: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 5. Render the HTMX HTML fragment
		tmpl, err := template.ParseFiles("web/templates/transactions.html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		data := struct {
			Transactions []models.Transaction
			Warnings     []string
		}{
			Transactions: statement.Transactions,
			Warnings:     statement.Warnings,
		}

		tmpl.Execute(w, data)
	}
}

// PageHandler serves the initial dashboard UI
func PageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
