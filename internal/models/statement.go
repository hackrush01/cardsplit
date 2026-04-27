package models

import "time"

// Statement represents the full credit card statement metadata and its transactions.
type Statement struct {
	CardType       string
	PaymentDueDate time.Time
	StatementDate  time.Time
	TotalAmountDue int
	MinAmountDue   int
	Transactions   []Transaction
	Warnings       []string
}
