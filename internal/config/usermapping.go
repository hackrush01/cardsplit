package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hackrush01/cardsplit/internal/models"
)

// UserCardMapping holds the parsed mapping data for efficient lookups.
type UserCardMapping struct {
	mapping   map[string]map[string]models.UserConfig // cardType -> cardHolderName -> UserConfig
	usernames []string
}

func MappingFilePath() string {
	p := os.Getenv("CONFIG_PATH")
	if p == "" {
		p = "./configs/user_card_mapping.json"
	}
	return p
}

// LoadCardMapping reads the JSON file and initializes the UserCardMapping structure.
func LoadCardMapping(filepath string) (*UserCardMapping, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("read mapping file at %v: %w", filepath, err)
	}

	var raw models.CardMappingFile
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal mapping file: %w", err)
	}

	cm := &UserCardMapping{
		mapping: make(map[string]map[string]models.UserConfig),
	}

	userSet := make(map[string]bool)
	for userKey, cards := range raw.Users {
		userSet[userKey] = true
		for _, card := range cards {
			if err := cm.addMapping(userKey, card); err != nil {
				return nil, err
			}
		}
	}

	for u := range userSet {
		cm.usernames = append(cm.usernames, u)
	}

	return cm, nil
}

func (cm *UserCardMapping) addMapping(username string, card models.CardConfig) error {
	if card.CardHolderName == "" {
		return fmt.Errorf("card_holder_name is required for user %q", username)
	}

	if _, exists := cm.mapping[card.CardType]; !exists {
		cm.mapping[card.CardType] = make(map[string]models.UserConfig)
	}

	if existing, exists := cm.mapping[card.CardType][card.CardHolderName]; exists {
		if existing.Username != username {
			return fmt.Errorf("card_holder_name %q is already mapped to a different user", card.CardHolderName)
		}
		return nil
	}

	cm.mapping[card.CardType][card.CardHolderName] = models.UserConfig{
		Username: username,
		CardType: card.CardType,
		Suffix:   card.Suffix,
	}
	return nil
}

// GetUserDetails retrieves the username and card suffix based on cardType and cardHolderName.
func (cm *UserCardMapping) GetUserDetails(cardType, cardHolderName string) (string, string, error) {
	if cardType == "" || cardHolderName == "" {
		return "", "", fmt.Errorf("cardType and cardHolderName must be provided")
	}

	if cardMap, exists := cm.mapping[cardType]; exists {
		if userConfig, exists := cardMap[cardHolderName]; exists {
			return userConfig.Username, userConfig.Suffix, nil
		}
	}

	return "", "", fmt.Errorf("no mapping found for cardType %q and cardHolderName %q", cardType, cardHolderName)
}

// Username returns the list of all configured usernames, including the injected "Admin" user.
func (cm *UserCardMapping) Usernames() []string {
	return cm.usernames
}

// GetConfiguredUsers is a helper to load the mapping and return configured usernames.
// This maintains compatibility with the existing API while using the updated models.
func GetConfiguredUsers(mappingFilePath string) ([]string, error) {
	cm, err := LoadCardMapping(mappingFilePath)
	if err != nil {
		return nil, err
	}
	return cm.Usernames(), nil
}
