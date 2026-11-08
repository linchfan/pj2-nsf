package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type sqliteDb struct {
	db *sql.DB
}

func newSqliteDb(dbPath string) (*sqliteDb, error) {
	// Open the SQLite database.
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Create the services table if it does not exist.
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS services (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to create services table: %w; failed to close database: %v", err, closeErr)
		}
		return nil, fmt.Errorf("failed to create services table: %w", err)
	}

	return &sqliteDb{db: conn}, nil
}

func (s *sqliteDb) Release() error {
	return s.db.Close()
}

func (s *sqliteDb) Exist(key string) (bool, error) {
	var exist bool

	// Query the row by key.
	row := s.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM services WHERE key = ?)",
		key,
	)

	// Check whether the key exists in the services table.
	err := row.Scan(&exist)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func (s *sqliteDb) Save(key, value string) error {
	// Insert the key and value into the table service.
	// Update the value if the key already exists.
	_, err := s.db.Exec(`
			INSERT INTO services (key, value)
			VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value
		`, key, value)
	if err != nil {
		return fmt.Errorf("failed to save key %s: %w", key, err)
	}

	return nil
}

func (s *sqliteDb) SaveAll(data map[string]string) error {
	// Begin a transaction for atomic batch insertion.
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Insert the key-value pairs into the services table.
	// Update the value if the key already exists.
	for key, value := range data {
		_, err := tx.Exec(`
			INSERT INTO services (key, value)
			VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value
		`, key, value)
		if err != nil {
			return fmt.Errorf("failed to save key %s: %w", key, err)
		}
	}

	// Commit all changes if every insertion succeeds.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *sqliteDb) Load(key string) (string, error) {
	// Query the row by key.
	row := s.db.QueryRow(
		"SELECT value FROM services WHERE key = ?;",
		key,
	)

	// Scan the value from the row.
	var value string
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to load key %s: %w", key, err)
	}

	return value, nil
}

func (s *sqliteDb) LoadAll() (map[string]string, error) {
	result := make(map[string]string)

	// Query all key-value pairs from the services table.
	rows, err := s.db.Query("SELECT key, value FROM services;")
	if err != nil {
		return nil, err
	}

	// Ensure that the query resource will be released.
	defer func() {
		_ = rows.Close()
	}()

	// Scan the keys and values from the rows.
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *sqliteDb) Remove(key string) error {
	// Remove the specified key from the services table.
	_, err := s.db.Exec(
		"DELETE FROM services WHERE key = ?",
		key,
	)
	if err != nil {
		return fmt.Errorf("failed to remove key %s: %w", key, err)
	}

	return nil
}

func (s *sqliteDb) RemoveAll() error {
	// Remove all key-value pairs from the services table.
	_, err := s.db.Exec("DELETE FROM services;")

	if err != nil {
		return fmt.Errorf("failed to remove all services: %w", err)
	}
	return nil
}
