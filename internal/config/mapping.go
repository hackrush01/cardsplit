package config

import (
	"encoding/json"
	"os"
)

// UserConfig represents the JSON structure of a single mapped user
type UserConfig struct {
	Name     string `json:"name"`
	CardType string `json:"card_type"`
	Suffix   string `json:"suffix"`
}

// LoadCardMapping reads the JSON file and returns a map for O(1) memory lookups
func LoadCardMapping(filepath string) (map[string]UserConfig, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var mapping map[string]UserConfig
	if err := json.Unmarshal(bytes, &mapping); err != nil {
		return nil, err
	}

	return mapping, nil
}
