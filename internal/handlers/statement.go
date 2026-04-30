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
		// Get logged-in user
		username := r.Context().Value(middleware.UsernameKey).(string)
		fmt.Printf("StatementViewHandler: User '%s' accessed the statement page\n", username)

		// 1. Read selections from the URL query
		selectedCard := r.URL.Query().Get("card_name")
		selectedDate := r.URL.Query().Get("statement_date")

		// 2. Fetch dropdown options for the UI
		// Note: You will need to ensure these DB methods exist in your storage package
		availableCards, _ := storage.CardsByUser(db, username)
		fmt.Printf("Available cards for user '%s': %v\n", username, availableCards)

		var availableDates []string
		if selectedCard != "" {
			// Only fetch dates if a card is selected, cascading the dropdowns
			availableDates, _ = storage.StatementDates(db, username, selectedCard)
		}
		fmt.Printf("Available statement dates for card '%s': %v\n", selectedCard, availableDates)

		// 3. Fetch data if selections are complete
		var transactions []models.Transaction
		var totalDuePaise int

		if selectedCard != "" && selectedDate != "" {
			transactions, _ = storage.TransactionsByStatement(db, username, selectedCard, selectedDate)

			// Calculate the summary
			for _, t := range transactions {
				totalDuePaise += t.Amount
			}
		}

		// 4. Prepare data for the template
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

		// 5. Parse BOTH templates so the {{ template }} inclusion works
		tmplFiles := []string{
			filepath.Join("web", "templates", "statement.html"),
			filepath.Join("web", "templates", "transactions.html"),
		}

		tmpl, err := template.ParseFiles(tmplFiles...)
		if err != nil {
			http.Error(w, "Failed to load templates", http.StatusInternalServerError)
			return
		}

		// Execute the specific statement template
		tmpl.ExecuteTemplate(w, "statement.html", data)
	}
}
