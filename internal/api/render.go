package api

import (
	"database/sql"
	"net/http"

	"github.com/hackrush01/cardsplit/internal/handlers"
	"github.com/hackrush01/cardsplit/internal/models"
	"github.com/hackrush01/cardsplit/internal/storage"
)

// renderTransactionsFragment fetches the latest transactions and renders the transaction list and adjustment form fragments.
func renderTransactionsFragment(w http.ResponseWriter, db *sql.DB, cardType, statementDate string, warnings []string) {
	csvTxns, manualTxns, err := storage.GetStatementTransactions(db, cardType, statementDate)
	if err != nil {
		http.Error(w, "Failed to load transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	users, err := storage.GetAllUsers(db, true)
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
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
	}{
		Transactions:       csvTxns,
		Warnings:           warnings,
		ManualTransactions: manualTxns,
		Users:              users,
		CardType:           cardType,
		StatementDate:      statementDate,
		TotalRewards:       totalRewards,
	}

	tmplFiles := []string{"web/templates/transactions.html", "web/templates/admin_dashboard.html"}
	handlers.RenderTemplate(w, data, tmplFiles...)
	handlers.RenderTemplateName(w, "adjustment-form", data, tmplFiles...)
}
