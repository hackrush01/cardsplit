package parsers

import (
	"strconv"

	"github.com/hackrush01/cardsplit/internal/config"
	"github.com/hackrush01/cardsplit/internal/models"
)

// enrichTransaction applies the card rules to infer IsPayment and RewardMultiplier
func enrichTransaction(cardType string, tx *models.Transaction) {
	rules, exists := config.CompiledRules[cardType]
	if !exists {
		return // No rules configured for this card
	}

	for _, re := range rules.Payments {
		if re.MatchString(tx.Description) {
			tx.IsPayment = true
			break
		}
	}

	for multiplier, regexList := range rules.Rewards {
		for _, re := range regexList {
			if re.MatchString(tx.Description) {
				tx.RewardMultiplier, _ = strconv.Atoi(multiplier)
				return
			}
		}
	}
	tx.RewardMultiplier = 1
}
