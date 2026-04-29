package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// CardConfig represents the JSON fields for a single card entry.
type CardConfig struct {
	CardType string `json:"card_type"`
	Suffix   string `json:"suffix"`
	CSVName  string `json:"csv_name"`
}

// UserConfig represents the flattened lookup data used by the app.
type UserConfig struct {
	Username string `json:"username"`
	CardType string `json:"card_type"`
	Suffix   string `json:"suffix"`
}

// cardMappingFile is the structure of the new card mapping JSON.
type cardMappingFile struct {
	Users map[string][]CardConfig `json:"users"`
}

// CardMapping holds the parsed mapping data for efficient lookups.
type CardMapping struct {
	mapping map[string]map[string]UserConfig // cardType -> csvName -> UserConfig
}

// LoadCardMapping reads the JSON file and initializes the CardMapping structure.
func LoadCardMapping(filepath string) (*CardMapping, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var raw cardMappingFile
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, err
	}

	mapping := make(map[string]map[string]UserConfig)
	for userKey, cards := range raw.Users {
		for _, card := range cards {
			if card.CSVName == "" {
				return nil, fmt.Errorf("csv_name is required for user %q", userKey)
			}

			if _, exists := mapping[card.CardType]; !exists {
				mapping[card.CardType] = make(map[string]UserConfig)
			}

			if existing, exists := mapping[card.CardType][card.CSVName]; exists {
				if existing.Username != userKey {
					return nil, fmt.Errorf("csv_name %q is already mapped to a different user", card.CSVName)
				}
				continue
			}

			mapping[card.CardType][card.CSVName] = UserConfig{
				Username: userKey,
				CardType: card.CardType,
				Suffix:   card.Suffix,
			}
		}
	}

	return &CardMapping{mapping: mapping}, nil
}

// GetUserDetails retrieves the username and card suffix based on cardType and csvName.
func (cm *CardMapping) GetUserDetails(cardType, csvName string) (string, string, error) {
	if cardType == "" || csvName == "" {
		return "", "", fmt.Errorf("cardType and csvName must be provided")
	}

	if cardMap, exists := cm.mapping[cardType]; exists {
		if userConfig, exists := cardMap[csvName]; exists {
			return userConfig.Username, userConfig.Suffix, nil
		}
	}

	return "", "", fmt.Errorf("no mapping found for cardType %q and csvName %q", cardType, csvName)
}
