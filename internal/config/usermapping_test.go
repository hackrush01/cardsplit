package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCardMapping_AllowsDuplicateCardHolderNameForSameUser(t *testing.T) {
	contents := `{
	  "users": {
	    "alice": [
	      {"card_type": "Infinia", "suffix": "1234", "card_holder_name": "Alice"},
	      {"card_type": "Emeralde", "suffix": "9876", "card_holder_name": "Alice"}
	    ]
	  }
	}`
	tmpFile := filepath.Join(os.TempDir(), "card_mapping_test.json")
	if err := os.WriteFile(tmpFile, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	mapping, err := LoadCardMapping(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user, ok := mapping.mapping["Infinia"]["Alice"]
	if !ok {
		t.Fatalf("expected mapping for Alice")
	}
	if user.Username != "alice" {
		t.Fatalf("expected username alice, got %q", user.Username)
	}
}

func TestLoadCardMapping_DuplicateCardHolderNameAcrossUsersFails(t *testing.T) {
	contents := `{
	  "users": {
	    "alice": [
	      {"card_type": "Infinia", "suffix": "1234", "card_holder_name": "Alice"}
	    ],
	    "bob": [
	      {"card_type": "Infinia", "suffix": "5678", "card_holder_name": "Alice"}
	    ]
	  }
	}`
	tmpFile := filepath.Join(os.TempDir(), "card_mapping_dupe_test.json")
	if err := os.WriteFile(tmpFile, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	if _, err := LoadCardMapping(tmpFile); err == nil {
		t.Fatal("expected error for duplicate card_holder_name across users")
	}
}

func TestGetUserDetails(t *testing.T) {
	contents := `{
	  "users": {
	    "alice": [
	      {"card_type": "Infinia", "suffix": "1234", "card_holder_name": "Alice"}
	    ]
	  }
	}`
	tmpFile := filepath.Join(os.TempDir(), "card_mapping_get_user_test.json")
	if err := os.WriteFile(tmpFile, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	mapping, err := LoadCardMapping(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	username, suffix, err := mapping.GetUserDetails("Infinia", "Alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if username != "alice" {
		t.Fatalf("expected username alice, got %q", username)
	}
	if suffix != "1234" {
		t.Fatalf("expected suffix 1234, got %q", suffix)
	}

	_, _, err = mapping.GetUserDetails("Infinia", "NonExistent")
	if err == nil {
		t.Fatal("expected error for non-existent mapping")
	}
}
