package models

import "time"

// Statement represents the full credit card statement metadata and its transactions.
type Statement struct {
	CardType       string
	StatementDate  time.Time
	PaymentDueDate time.Time
	TotalAmountDue int
	MinAmountDue   int
	Warnings       []string
	Transactions   []Transaction
}
