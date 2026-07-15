package db

import (
	_ "embed"
	"fmt"

	"database/sql"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// Open opens the SQLite database at path, enables WAL mode, runs migrations,
// and ensures the singleton account row exists with balance 0.
func Open(path string) (*sql.DB, error) {
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := conn.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}

	if _, err := conn.Exec(`INSERT OR IGNORE INTO accounts (id, balance) VALUES (1, 0)`); err != nil {
		return nil, fmt.Errorf("seed account: %w", err)
	}

	return conn, nil
}
