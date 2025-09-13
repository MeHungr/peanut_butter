package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	DB *sqlx.DB
}

func NewStorage(path string) (*Storage, error) {
	// Open a connection to the DB
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("Failed to open SQLite connection: %w", err)
	}

	// Ping the DB to ensure it's working
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to connect to DB: %w", err)
	}

	// Wrap the database in the storage struct and return it
	storage := &Storage{DB: db}
	return storage, nil
}
