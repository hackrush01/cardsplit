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
		SELECT transaction_type, transaction_timestamp, description, amount, base_reward_value, reward_multiplier, is_payment, is_manual, is_settlement
		FROM transactions 
		WHERE username = ? AND card_type = ? AND statement_date = ? 
		ORDER BY is_manual DESC,transaction_timestamp ASC`,
		username, cardType, statementDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.TxnType,
			&t.TxnTimestamp,
			&t.Description,
			&t.Amount,
			&t.BaseRewardValue,
			&t.RewardMultiplier,
			&t.IsPayment,
			&t.IsManual,
			&t.IsSettlement,
		)
		if err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	return txs, nil
}

// GetStatementTransactions fetches both CSV and manual transactions for a specific statement in one query.
func GetStatementTransactions(db *sql.DB, cardType string, statementDate string) (csv, manual []models.Transaction, err error) {
	rows, err := db.Query(`
		SELECT username, transaction_type, transaction_timestamp, description, card_holder_name, amount, base_reward_value, reward_multiplier, is_payment, is_manual, is_settlement
		FROM transactions
		WHERE card_type = ? AND statement_date = ?
		ORDER BY transaction_timestamp DESC`,
		cardType, statementDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.Username,
			&t.TxnType,
			&t.TxnTimestamp,
			&t.Description,
			&t.CardHolderName,
			&t.Amount,
			&t.BaseRewardValue,
			&t.RewardMultiplier,
			&t.IsPayment,
			&t.IsManual,
			&t.IsSettlement,
		); err != nil {
			return nil, nil, err
		}
		if t.IsManual {
			manual = append(manual, t)
		} else {
			csv = append(csv, t)
		}
	}
	return csv, manual, nil
}

// AddManualTransaction inserts a manually created transaction.
func AddManualTransaction(db *sql.DB, cardType, statementDate string, t models.Transaction) error {
	_, err := db.Exec(`
		INSERT INTO transactions (card_type, statement_date, key_timestamp, username, transaction_type, transaction_timestamp, card_holder_name, description, amount, is_manual, is_settlement)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0)`,
		cardType,
		statementDate,
		t.KeyTimestamp.Format("2006-01-02 15:04:05"),
		t.Username,
		t.TxnType,
		t.TxnTimestamp.Format("2006-01-02 15:04:05"),
		t.CardHolderName,
		t.Description,
		t.Amount,
	)
	return err
}

// GetPendingBalances returns the pending amount for each card type for a given user.
func GetPendingBalances(db *sql.DB, username string) (map[string]int, error) {
	rows, err := db.Query("SELECT card_type, COALESCE(SUM(amount), 0) FROM transactions WHERE username = ? AND is_payment = 0 GROUP BY card_type", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make(map[string]int)
	for rows.Next() {
		var cardType string
		var amount int
		if err := rows.Scan(&cardType, &amount); err != nil {
			return nil, err
		}
		balances[cardType] = amount
	}
	return balances, nil
}

// SettleCardDues adds a settlement transaction for a specific card to zero out its pending total.
func SettleCardDues(tx *sql.Tx, username, cardType, cardHolderName string, amount int) error {
	var statementDateTime time.Time
	err := tx.QueryRow(`
		SELECT statement_date 
		FROM transactions 
		WHERE username = ? AND card_type = ?
		ORDER BY statement_date DESC LIMIT 1`, username, cardType).Scan(&statementDateTime)
	if err != nil {
		return err
	}
	statementDate := statementDateTime.Format("2006-01-02")

	t := models.Transaction{
		KeyTimestamp:   time.Now(),
		Username:       username,
		TxnType:        "Settle Dues",
		TxnTimestamp:   time.Now(),
		CardHolderName: cardHolderName,
		Description:    "Admin marked dues as paid",
		Amount:         -amount,
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (card_type, statement_date, key_timestamp, username, transaction_type, transaction_timestamp, card_holder_name, description, amount, is_manual, is_payment, is_settlement)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0, 1)`,
		cardType,
		statementDate,
		t.KeyTimestamp.Format("2006-01-02 15:04:05"),
		t.Username,
		t.TxnType,
		t.TxnTimestamp.Format("2006-01-02 15:04:05"),
		t.CardHolderName,
		t.Description,
		t.Amount,
	)
	return err
}
