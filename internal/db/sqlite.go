package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const Driver = "sqlite3"

// SQLiteStore is a SQLite store
type SQLiteStore struct {
	*sql.DB
}

// Open opens a SQLite database
func Open(url string) (*SQLiteStore, error) {
	db, err := sql.Open(Driver, url)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{db}, nil
}

// IsSupportedDriver returns true if the driver is supported
func IsSupportedDriver(driver string) bool {
	return driver == "sqlite3"
}
