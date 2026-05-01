package storage

import (
	"database/sql"
	"encoding/json"

	"github.com/hackrush01/cardsplit/internal/models"
)

// SaveStatement atomically saves the statement metadata and all its transactions
func SaveStatement(db *sql.DB, stmt *models.Statement) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	warningsJSON, err := json.Marshal(stmt.Warnings)
	if err != nil {
		return err
	}

	// 1. Insert or Update the statement record
	_, err = tx.Exec(`
		INSERT INTO statements (card_type, statement_date, payment_due_date, total_amount_due, min_amount_due, warnings)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(card_type, statement_date) DO UPDATE SET
			payment_due_date = excluded.payment_due_date,
			total_amount_due = excluded.total_amount_due,
			min_amount_due = excluded.min_amount_due,
			warnings = excluded.warnings,
			uploaded_at = CURRENT_TIMESTAMP`,
		stmt.CardType,
		stmt.StatementDate.Format("2006-01-02"),
		stmt.PaymentDueDate.Format("2006-01-02"),
		stmt.TotalAmountDue,
		stmt.MinAmountDue,
		string(warningsJSON))
	if err != nil {
		return err
	}

	// 2. Delete existing transactions for this statement (to replace them)
	if _, err := tx.Exec("DELETE FROM transactions WHERE card_type = ? AND statement_date = ?", stmt.CardType, stmt.StatementDate.Format("2006-01-02")); err != nil {
		return err
	}

	// 3. Insert new transactions
	ins, err := tx.Prepare(`
		INSERT INTO transactions (card_type, statement_date, key_timestamp, username, transaction_timestamp, card_holder_name, description, amount, base_reward_value, reward_multiplier)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for _, t := range stmt.Transactions {
		_, err := ins.Exec(
			stmt.CardType,
			stmt.StatementDate.Format("2006-01-02"),
			t.KeyTimestamp.Format("2006-01-02 15:04:05"),
			t.Username,
			t.TxnTimestamp.Format("2006-01-02 15:04:05"),
			t.CardHolderName,
			t.Description,
			t.Amount,
			t.BaseRewardValue,
			t.RewardMultiplier,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
