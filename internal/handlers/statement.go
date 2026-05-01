package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/hackrush01/cardsplit/internal/middleware"
	"github.com/hackrush01/cardsplit/internal/models"
	"github.com/hackrush01/cardsplit/internal/storage"
)

func StatementViewHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(middleware.UsernameKey).(string)

		selectedCard := r.URL.Query().Get("card_name")
		selectedDate := r.URL.Query().Get("statement_date")

		availableCards, _ := storage.CardsByUser(db, username)
		fmt.Printf("Available cards for user '%s': %v\n", username, availableCards)

		var availableDates []string
		if selectedCard != "" {
			availableDates, _ = storage.StatementDates(db, username, selectedCard)
		}
		fmt.Printf("Available statement dates for card '%s': %v\n", selectedCard, availableDates)

		var transactions []models.Transaction
		var totalDuePaise int

		if selectedCard != "" && selectedDate != "" {
			transactions, _ = storage.TransactionsByStatement(db, username, selectedCard, selectedDate)

			for _, t := range transactions {
				if !t.IsPayment {
					totalDuePaise += t.Amount
				}
			}
		}

		data := struct {
			AvailableCards []string
			AvailableDates []string
			SelectedCard   string
			SelectedDate   string
			Transactions   []models.Transaction
			TotalDueRupees float64
		}{
			AvailableCards: availableCards,
			AvailableDates: availableDates,
			SelectedCard:   selectedCard,
			SelectedDate:   selectedDate,
			Transactions:   transactions,
			TotalDueRupees: float64(totalDuePaise) / 100.0,
		}

		tmplFiles := []string{
			filepath.Join("web", "templates", "statement.html"),
			filepath.Join("web", "templates", "transactions.html"),
		}

		tmpl, err := template.ParseFiles(tmplFiles...)
		if err != nil {
			http.Error(w, "Failed to load templates", http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "statement.html", data)
	}
}
