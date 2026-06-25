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

// MarkPaidHandler marks all dues as paid for a given user
func MarkPaidHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username := r.FormValue("username")
		if username == "" {
			http.Error(w, "Username required", http.StatusBadRequest)
			return
		}

		cm, err := config.LoadCardMapping(config.MappingFilePath())
		if err != nil {
			http.Error(w, "Load user configurations", http.StatusInternalServerError)
			return
		}

		balances, err := storage.GetPendingBalances(db, username)
		if err != nil {
			http.Error(w, "Failed to get pending balances: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		for cardType, amount := range balances {
			if amount <= 0 {
				continue
			}
			chn, err := cm.GetCardHolderName(cardType, username)
			if err != nil {
				http.Error(w, "Could not find card holder name for user", http.StatusBadRequest)
				return
			}
			if err := storage.SettleCardDues(tx, username, cardType, chn, amount); err != nil {
				http.Error(w, "Failed to settle dues: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	}
}
