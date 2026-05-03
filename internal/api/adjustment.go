package api

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/hackrush01/cardsplit/internal/config"
	"github.com/hackrush01/cardsplit/internal/models"
	"github.com/hackrush01/cardsplit/internal/storage"
)

// AdjustmentHandler adds a manual adjustment to a statement
func AdjustmentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cardType := r.FormValue("card_type")
		statementDate := r.FormValue("statement_date")
		username := r.FormValue("username")
		description := r.FormValue("description")
		amountStr := r.FormValue("amount")

		parsedAmount, err := strconv.Atoi(amountStr)
		if err != nil {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}

		cm, err := config.LoadCardMapping(config.MappingFilePath())
		if err != nil {
			http.Error(w, "Load user configurations", http.StatusInternalServerError)
			return
		}

		chn, err := cm.GetCardHolderName(cardType, username)
		if err != nil {
			http.Error(w, "Could not find card holder name for user", http.StatusBadRequest)
			return
		}

		txn := models.Transaction{
			KeyTimestamp:   time.Now(),
			Username:       username,
			TxnType:        "Manual adjustment",
			TxnTimestamp:   time.Now(),
			CardHolderName: chn,
			Description:    description,
			Amount:         parsedAmount * 100,
		}

		if err := storage.AddManualTransaction(db, cardType, statementDate, txn); err != nil {
			http.Error(w, "Failed to save adjustment: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Reload everything to re-render the fragment
		csvTxns, manualTxns, err := storage.GetStatementTransactions(db, cardType, statementDate)
		if err != nil {
			http.Error(w, "Load transactions", http.StatusInternalServerError)
			return
		}
		users, err := storage.GetAllUsers(db, true)
		if err != nil {
			http.Error(w, "Load users", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("web/templates/transactions.html", "web/templates/admin_dashboard.html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		totalRewards := 0
		for _, t := range csvTxns {
			totalRewards += t.TotalRewards()
		}
		for _, t := range manualTxns {
			totalRewards += t.TotalRewards()
		}

		data := struct {
			Transactions       []models.Transaction
			Warnings           []string
			ManualTransactions []models.Transaction
			Users              []string
			CardType           string
			StatementDate      string
			TotalRewards       int
		}{csvTxns, nil, manualTxns, users, cardType, statementDate, totalRewards}

		if err := tmpl.Execute(w, data); err != nil {
			return
		}
		tmpl.ExecuteTemplate(w, "adjustment-form", data)
	}
}
