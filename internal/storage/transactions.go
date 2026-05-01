package storage

import (
	"database/sql"
	"time"

	"github.com/hackrush01/cardsplit/internal/models"
)

// CardsByUser returns a list of unique card types associated with a specific user.
func CardsByUser(db *sql.DB, username string) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT card_type FROM transactions WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []string
	for rows.Next() {
		var card string
		if err := rows.Scan(&card); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// StatementDates returns a list of unique statement dates for a specific card and user.
func StatementDates(db *sql.DB, username string, cardType string) ([]string, error) {
	rows, err := db.Query(`
		SELECT DISTINCT statement_date 
		FROM transactions 
		WHERE username = ? AND card_type = ? 
		ORDER BY statement_date DESC`,
		username, cardType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date.Format("2006-01-02"))
	}
	return dates, nil
}

// TransactionsByStatement retrieves all transactions for a given user's statement.
func TransactionsByStatement(db *sql.DB, username string, cardType string, statementDate string) ([]models.Transaction, error) {
	rows, err := db.Query(`
		SELECT card_type, transaction_timestamp, description, amount, base_reward_value, reward_multiplier
		FROM transactions 
		WHERE username = ? AND card_type = ? AND statement_date = ? 
		ORDER BY transaction_timestamp ASC`,
		username, cardType, statementDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.Type,
			&t.TxnTimestamp,
			&t.Description,
			&t.Amount,
			&t.BaseRewardValue,
			&t.RewardMultiplier,
		)
		if err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	return txs, nil
}
