package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

type CardRulesRaw struct {
	Cards map[string]struct {
		Payments []string            `json:"payments"`
		Rewards  map[string][]string `json:"rewards"`
	} `json:"cards"`
}

type CardRuleCompiled struct {
	Payments []*regexp.Regexp
	Rewards  map[string][]*regexp.Regexp // Multiplier -> Slice of compiled regexes
}

func RuleFilePath() string {
	crp := os.Getenv("CARD_RULES_PATH")
	if crp == "" {
		crp = "./configs/card_rules.json"
	}
	return crp
}

var CompiledRules map[string]*CardRuleCompiled

func LoadCardRules(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read card rules config: %w", err)
	}

	var rawConfig CardRulesRaw
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("failed to parse card rules JSON: %w", err)
	}

	CompiledRules = make(map[string]*CardRuleCompiled)

	for cardName, cardData := range rawConfig.Cards {
		compiled := &CardRuleCompiled{
			Rewards: make(map[string][]*regexp.Regexp),
		}

		// Compile Payment Regexes
		for _, p := range cardData.Payments {
			re, err := regexp.Compile(p)
			if err != nil {
				return fmt.Errorf("invalid regex for %s payment: %w", cardName, err)
			}
			compiled.Payments = append(compiled.Payments, re)
		}

		// Compile Reward Regexes
		for multiplier, regexes := range cardData.Rewards {
			for _, r := range regexes {
				re, err := regexp.Compile(r)
				if err != nil {
					return fmt.Errorf("invalid regex for %s reward %s: %w", cardName, multiplier, err)
				}
				compiled.Rewards[multiplier] = append(compiled.Rewards[multiplier], re)
			}
		}

		CompiledRules[cardName] = compiled
	}

	return nil
}
