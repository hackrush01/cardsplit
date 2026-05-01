package models

import "time"

// Transaction represents a standardized ledger entry, regardless of the bank.
type Transaction struct {
	Type             string
	KeyTimestamp     time.Time // key timestamp is used for storage uniqueness
	Username         string
	TxnTimestamp     time.Time // real transaction timestamp
	CardHolderName   string
	Description      string
	Amount           int
	BaseRewardValue  int
	RewardMultiplier int
}

// AmountFloat is a helper method used exclusively by the UI to display the float value
func (t Transaction) AmountFloat() float64 {
	return float64(t.Amount) / 100.0
}

// TotalRewards calculates the final reward points
func (t Transaction) TotalRewards() int {
	return int(t.BaseRewardValue * t.RewardMultiplier)
}
