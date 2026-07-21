// internal/services/db.go

package services

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/3rr0r-505/Presenz/internal/config"
	"github.com/3rr0r-505/Presenz/internal/models"
)

type DB struct {
	conn *sql.DB
}

func Connect(path string, cfg *config.Config) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("database not found: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	if cfg.Database.WalMode {
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			return nil, fmt.Errorf("enabling WAL mode: %w", err)
		}
	}

	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d;", cfg.Database.TimeoutSeconds*1000)); err != nil {
		return nil, fmt.Errorf("setting busy timeout: %w", err)
	}

	return &DB{conn: db}, nil
}

// CreateTable creates the attendance table for a session if it doesn't exist.
// tableName must already be sanitized upstream (session service).
func (d *DB) CreateTable(tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s" (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			roll TEXT NOT NULL UNIQUE,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`, tableName)

	if _, err := d.conn.Exec(query); err != nil {
		return fmt.Errorf("creating table %q: %w", tableName, err)
	}
	return nil
}

// InsertAttendance inserts a record. Returns the raw driver error so the
// caller (routes layer) can detect UNIQUE constraint violations on roll.
func (d *DB) InsertAttendance(tableName, name, roll string) error {
	query := fmt.Sprintf(`INSERT INTO "%s" (name, roll) VALUES (?, ?);`, tableName)

	if _, err := d.conn.Exec(query, name, roll); err != nil {
		return fmt.Errorf("inserting attendance: %w", err)
	}
	return nil
}

// FetchAll returns all records for a session table, used by export on exit.
func (d *DB) FetchAll(tableName string) ([]models.AttendanceEntry, error) {
	query := fmt.Sprintf(`SELECT name, roll, timestamp FROM "%s";`, tableName)

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("fetching records from %q: %w", tableName, err)
	}
	defer rows.Close()

	var records []models.AttendanceEntry
	for rows.Next() {
		var e models.AttendanceEntry
		if err := rows.Scan(&e.Name, &e.Roll, &e.Timestamp); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		records = append(records, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return records, nil
}

// Close closes the underlying connection pool.
func (d *DB) Close() error {
	if d.conn == nil {
		return nil
	}
	return d.conn.Close()
}
