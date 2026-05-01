package api

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/hackrush01/cardsplit/internal/config"
	"github.com/hackrush01/cardsplit/internal/models"
	"github.com/hackrush01/cardsplit/internal/parsers"
	"github.com/hackrush01/cardsplit/internal/storage"
)

// UploadHandler processes the HTMX form submission
func UploadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		cardMapping, err := config.LoadCardMapping(config.MappingFilePath())
		if err != nil {
			http.Error(w, "Load user configurations", http.StatusInternalServerError)
			return
		}

		stmt, err := parsers.ParseInfiniaCSV(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<div class="rounded border border-red-300 bg-red-50 p-4 text-red-700"><strong>Error:</strong> %s</div>`, template.HTMLEscapeString(err.Error()))
			return
		}

		for i, txn := range stmt.Transactions {
			u, _, err := cardMapping.GetUserDetails("Infinia", txn.CardHolderName)
			if err == nil {
				stmt.Transactions[i].Username = u
			} else {
				http.Error(w, "Unrecognized cardholder name: "+txn.CardHolderName, http.StatusBadRequest)
				return
			}
		}

		if err := storage.SaveStatement(db, stmt); err != nil {
			http.Error(w, "Failed to save statement: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("web/templates/transactions.html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		data := struct {
			Transactions []models.Transaction
			Warnings     []string
		}{
			Transactions: stmt.Transactions,
			Warnings:     stmt.Warnings,
		}

		tmpl.Execute(w, data)
	}
}
