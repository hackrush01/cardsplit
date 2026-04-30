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

	// Bootstrap admin acccont
	_, err = db.Exec(`INSERT OR IGNORE INTO users (username, role) VALUES (?, ?)`, "Admin", "admin")
	if err != nil {
		log.Fatalf("Inserting admin user: %v", err)
	}

	return db
}

// EnsureUsersExist checks if users are in the DB and creates them if they aren't.
func EnsureUsersExist(users []string, db *sql.DB) error {
	stmt, err := db.Prepare(`INSERT OR IGNORE INTO users (username) VALUES (?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, user := range users {
		if _, err := stmt.Exec(user); err != nil {
			log.Printf("Inserting user %s: %v", user, err)
			return err
		}
	}

	return nil
}

// GetAllUsers fetches all usernames for the login dropdown.
func GetAllUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT username FROM users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func GetUserRole(db *sql.DB, username string) string {
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE username = ?", username).Scan(&role)
	if err != nil {
		log.Fatalf("Unable to determine user role.")
	}
	return role
}

func createSchema(db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NULL,
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
