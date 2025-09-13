package storage

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	// "github.com/mattn/go-sqlite3"
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
	status TEXT,
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
	FOREIGN KEY (task_id) REFERENCES tasks(task_id),
	FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);
`

	_, err := s.DB.Exec(schema)
	return err
}

func (s *Storage) RegisterAgent(agent Agent) error {
	query := `
INSERT INTO agents (agent_id, os, arch, targeted, agent_ip, server_ip, server_port, callback_interval, hostname, status, last_seen)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(agent_id) DO UPDATE SET
    os = excluded.os,
    arch = excluded.arch,
    targeted = excluded.targeted,
    agent_ip = excluded.agent_ip,
    server_ip = excluded.server_ip,
    server_port = excluded.server_port,
    callback_interval = excluded.callback_interval,
    hostname = excluded.hostname,
    status = excluded.status,
    last_seen = excluded.last_seen;
`
	if _, err := s.DB.Exec(query,
		agent.ID, agent.OS, agent.Arch, agent.Targeted,
		agent.AgentIP, agent.ServerIP, agent.ServerPort,
		agent.CallbackInterval, agent.Hostname, agent.Status,
		agent.LastSeen,
	); err != nil {
		return fmt.Errorf("RegisterAgent: %w", err)
	}

	return nil
}
