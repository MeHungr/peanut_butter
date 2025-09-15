// Contains all database logic
package storage

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	DB *sqlx.DB
}

// NewStorage creates a new Storage that contains a DB
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

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("Failed to enable foreign keys: %w", err)
	}

	// Wrap the database in the storage struct, initialize the schema, and return it
	storage := &Storage{DB: db}
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("Failed to initialize schema: %w", err)
	}
	log.Println("Database schema initialized successfully")
	return storage, nil
}

// initSchema initializes the db with proper tables
func (s *Storage) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS agents (
	agent_id TEXT PRIMARY KEY,
	os TEXT,
	arch TEXT,
	targeted BOOLEAN,
	agent_ip TEXT,
	server_ip TEXT,
	server_port INTEGER,
	callback_interval INTEGER,
	hostname TEXT,
	last_seen TIMESTAMP
);
CREATE TABLE IF NOT EXISTS tasks (
	task_id INTEGER PRIMARY KEY AUTOINCREMENT,
	agent_id TEXT NOT NULL,
	type TEXT,
	completed BOOLEAN,
	payload TEXT,
	timeout INTEGER,
	timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);
CREATE TABLE IF NOT EXISTS results (
	result_id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	agent_id TEXT NOT NULL,
	output TEXT,
	return_code INTEGER,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (task_id, agent_id),
	FOREIGN KEY (task_id) REFERENCES tasks(task_id),
	FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);
CREATE INDEX IF NOT EXISTS idx_agents_id ON agents(agent_id);
CREATE INDEX IF NOT EXISTS idx_agents_targeted ON agents(targeted);
CREATE INDEX IF NOT EXISTS idx_tasks_agent_completed ON tasks(agent_id, completed);
CREATE INDEX IF NOT EXISTS idx_results_task ON results(task_id);
CREATE INDEX IF NOT EXISTS idx_results_agent ON results(agent_id);
`

	_, err := s.DB.Exec(schema)
	return err
}
