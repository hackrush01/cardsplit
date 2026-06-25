package api

import (
	"database/sql"
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

		renderTransactionsFragment(w, db, cardType, statementDate, nil)
	}
}
