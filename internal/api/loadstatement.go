package api

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/hackrush01/cardsplit/internal/storage"
)

// LoadStatementHandler loads an existing statement and returns its transactions
func LoadStatementHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cardType := r.FormValue("card_type")
		statementDate := r.FormValue("statement_date")

		if cardType == "" || statementDate == "" {
			http.Error(w, "Missing card_type or statement_date", http.StatusBadRequest)
			return
		}

		// Get transactions for this statement
		csvTxns, err := storage.GetTransactionsByType(db, cardType, statementDate, false)
		if err != nil {
			http.Error(w, "Load transactions: "+err.Error(), http.StatusInternalServerError)
			return
		}

		manualTxns, err := storage.GetTransactionsByType(db, cardType, statementDate, true)
		if err != nil {
			http.Error(w, "Load adjustments: "+err.Error(), http.StatusInternalServerError)
			return
		}

		users, err := storage.GetAllUsers(db, true)
		if err != nil {
			http.Error(w, "Load users: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse and render transactions template
		tmpl, err := template.ParseFiles("web/templates/transactions.html", "web/templates/admin_dashboard.html")
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Transactions       interface{}
			Warnings           []string
			ManualTransactions interface{}
			Users              []string
			CardType           string
			StatementDate      string
		}{csvTxns, nil, manualTxns, users, cardType, statementDate}

		// Execute the transactions template
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute adjustment form OOB swap
		if err := tmpl.ExecuteTemplate(w, "adjustment-form", data); err != nil {
			return
		}
	}
}

// StatementDatesOptionsHandler returns HTML <option> tags for statement dates
func StatementDatesOptionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardType := r.URL.Query().Get("card_type")
		fmt.Fprint(w, `<option value="">Select Statement Date...</option>`)
		if cardType == "" {
			return
		}

		// Using raw SQL here to find unique statement dates for the selected card
		rows, err := db.Query("SELECT DISTINCT statement_date FROM transactions WHERE card_type = ? ORDER BY statement_date DESC", cardType)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var date time.Time
			if err := rows.Scan(&date); err == nil {
				fmt.Fprintf(w, `<option value="%[1]s">%[1]s</option>`, date.Format("2006-01-02"))
			}
		}
	}
}
