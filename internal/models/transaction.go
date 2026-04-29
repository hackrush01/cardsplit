package models

import "time"

// RewardData encapsulates reward logic to make it extensible for future multipliers
type RewardData struct {
	BaseValue  int
	Multiplier int
}

// Total calculates the final reward points
func (r RewardData) Total() int {
	return int(r.BaseValue * r.Multiplier)
}

// Transaction represents a standardized ledger entry, regardless of the bank.
type Transaction struct {
	Type             string
	ActualTimestamp  time.Time // actual transaction timestamp
	ShiftedTimestamp time.Time // shifted timestamp used for storage uniqueness
	Description      string
	Amount           int
	Rewards          RewardData
	RawLabel         string
	Username         string
}

// AmountFloat is a helper method used exclusively by the UI to display the float value
func (t Transaction) AmountFloat() float64 {
	return float64(t.Amount) / 100.0
}
