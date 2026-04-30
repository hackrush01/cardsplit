package models

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

// CardMappingFile is the structure of the new card mapping JSON.
type CardMappingFile struct {
	Users map[string][]CardConfig `json:"users"`
}
