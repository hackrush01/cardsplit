package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens the SQLite connection, applies best-practice PRAGMAs, and creates tables.
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Open database: %v", err)
	}

	// Apply SQLite performance and concurrency best practices
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA busy_timeout = 10000;",
		"PRAGMA foreign_keys = ON;",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			log.Fatalf("Pragma execute %s: %v", pragma, err)
		}
	}

	createSchema(db)
	return db
}

func createSchema(db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'basic',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS statements (
		card_type TEXT NOT NULL,
		statement_date DATE NOT NULL,
		payment_due_date DATE NOT NULL,
		total_amount_due INTEGER NOT NULL,
		min_amount_due INTEGER NOT NULL,
		warnings TEXT DEFAULT '[]',
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (card_type, statement_date)
	);

	CREATE TABLE IF NOT EXISTS transactions (
		card_type TEXT NOT NULL,
		statement_date DATE NOT NULL,
		transaction_timestamp DATETIME NOT NULL,
		actual_transaction_timestamp DATETIME NOT NULL,
		username TEXT,
		card_holder_name TEXT NOT NULL,
		description TEXT NOT NULL,
		amount INTEGER NOT NULL,
		rewards INTEGER DEFAULT 0,
		is_manual BOOLEAN DEFAULT 0,
		PRIMARY KEY (card_type, statement_date, transaction_timestamp),
		FOREIGN KEY (card_type, statement_date) REFERENCES statements(card_type, statement_date) ON DELETE CASCADE
		FOREIGN KEY (username) REFERENCES users(username) ON DELETE SET NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
}
