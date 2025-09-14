// Contains all database logic
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

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

// RegisterAgent handles registering an agent in the db
func (s *Storage) RegisterAgent(agent *Agent) error {
	// Query for the db
	query := `
INSERT INTO agents (agent_id, os, arch, targeted, agent_ip, server_ip, server_port, callback_interval, hostname, last_seen)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(agent_id) DO UPDATE SET
    os = excluded.os,
    arch = excluded.arch,
    targeted = excluded.targeted,
    agent_ip = excluded.agent_ip,
    server_ip = excluded.server_ip,
    server_port = excluded.server_port,
    callback_interval = excluded.callback_interval,
    hostname = excluded.hostname,
    last_seen = excluded.last_seen;
`
	// Execute the query and replace ? with each variable
	if _, err := s.DB.Exec(query,
		agent.ID, agent.OS, agent.Arch, agent.Targeted,
		agent.AgentIP, agent.ServerIP, agent.ServerPort,
		agent.CallbackInterval, agent.Hostname,
		agent.LastSeen,
	); err != nil {
		return fmt.Errorf("RegisterAgent: %w", err)
	}

	return nil
}

// GetAgentByID retrieves the agent with the ID passed in as an argument
func (s *Storage) GetAgentByID(agentID string) (*Agent, error) {
	// Query for the db
	query := `SELECT * FROM agents WHERE agent_id = ?`

	// Initialize agent struct
	var agent Agent
	// Query the db for the matching agent
	if err := s.DB.Get(&agent, query, agentID); err != nil {
		// If no agent is found, return nil
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		// Other errors get reported
		return nil, fmt.Errorf("GetAgentByID: %w", err)
	}

	// Return the pointer to the agent
	return &agent, nil
}

// UpdateLastSeen updates the last seen time of an agent to the provided time
func (s *Storage) UpdateLastSeen(agentID string, t time.Time) error {
	// Query for the db
	query := `UPDATE agents SET last_seen = ? WHERE agent_id = ?`

	// Execute the query
	if _, err := s.DB.Exec(query, t, agentID); err != nil {
		return fmt.Errorf("UpdateLastSeen: %w", err)
	}
	return nil
}

// GetNextTask returns the next task for the agent with agentID
func (s *Storage) GetNextTask(agentID string) (*Task, error) {
	// Query for the db
	query := `
SELECT * FROM tasks
WHERE agent_id = ? AND completed = 0
ORDER BY timestamp ASC
LIMIT 1
`

	// Initialize the new task
	var task Task
	// Query the db
	if err := s.DB.Get(&task, query, agentID); err != nil {
		// If no task is found, return nil
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		// Other errors get reported
		return nil, fmt.Errorf("GetNextTask: %w", err)
	}

	// Return the pointer to the task
	return &task, nil
}

// MarkTaskCompleted marks the task given by taskID as completed
func (s *Storage) MarkTaskCompleted(taskID int) error {
	// Query for the db
	query := `UPDATE tasks SET completed = 1 WHERE task_id = ?`

	// Update the task
	if _, err := s.DB.Exec(query, taskID); err != nil {
		return fmt.Errorf("MarkTaskCompleted: %w", err)
	}

	// No errors
	return nil
}

// InsertResult inserts a result into the db
func (s *Storage) InsertResult(r *Result) error {
	// Query for the db
	query := `
INSERT OR IGNORE INTO results (task_id, agent_id, output, return_code)
VALUES (?, ?, ?, ?)
`

	// Insert the result
	res, err := s.DB.Exec(query, r.TaskID, r.AgentID, r.Output, r.ReturnCode)
	if err != nil {
		return fmt.Errorf("InsertResult: %w", err)
	}

	// Detect duplicate results and log them
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("InsertResult: failed to get RowsAffected: %w", err)
	}
	if rows == 0 {
		log.Printf("Duplicate result ignored: agent=%s task=%d", r.AgentID, r.TaskID)
	}

	// Update the result struct's id to be the auto incremented one
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("InsertResult: failed to get LastInsertId: %w", err)
	}
	r.ResultID = int(id)

	return nil
}

// GetAllAgents returns a slice of all agents registered with the db
func (s *Storage) GetAllAgents() ([]Agent, error) {
	// Query for the db
	query := `SELECT * FROM agents`

	// Create slice of Agents and query the db to fill it
	var agents []Agent
	if err := s.DB.Select(&agents, query); err != nil {
		return nil, fmt.Errorf("GetAllAgents: %w", err)
	}

	return agents, nil
}

// AddTargets sets the passed in agents as targeted
func (s *Storage) AddTargets(agentIDs []string) error {
	// Prevents invalid SQL 'IN ()'
	if len(agentIDs) == 0 {
		return nil
	}

	// Query for the db; checks if agent id is in agentIDs
	query, args, err := sqlx.In(`UPDATE agents SET targeted = 1 WHERE agent_id IN (?)`, agentIDs)
	if err != nil {
		return fmt.Errorf("AddTargets: %w", err)
	}

	// Remaps the query to use SQLite formatting
	query = s.DB.Rebind(query)
	// Executes the query with args expanded to individual variables
	if _, err := s.DB.Exec(query, args...); err != nil {
		return fmt.Errorf("AddTargets: %w", err)
	}

	// No errors
	return nil
}

// ClearTargets sets all agents to untargeted
func (s *Storage) ClearTargets() error {
	// Query for the db
	query := `UPDATE agents SET targeted = 0`

	// Make all agents untargeted
	if _, err := s.DB.Exec(query); err != nil {
		return fmt.Errorf("ClearTargets: %w", err)
	}

	// No errors
	return nil
}

// Untarget sets the passed in agents as untargeted
func (s *Storage) Untarget(agentIDs []string) error {
	// Prevents invalid SQL 'IN ()'
	if len(agentIDs) == 0 {
		return nil
	}

	// Query for the db
	query, args, err := sqlx.In(`UPDATE agents SET targeted = 0 WHERE agent_id IN (?)`, agentIDs)
	if err != nil {
		return fmt.Errorf("Untarget: %w", err)
	}

	// Remaps the query to use SQLite formatting
	query = s.DB.Rebind(query)
	// Executes the query with args expanded to individual variables
	if _, err := s.DB.Exec(query, args...); err != nil {
		return fmt.Errorf("Untarget: %w", err)
	}

	// No errors
	return nil
}

// GetTargets returns a slice of all agents set to targeted in the db
func (s *Storage) GetTargets() ([]Agent, error) {
	// Query for the db
	query := `SELECT * FROM agents WHERE targeted = 1`

	// Create slice of Agents and query the db to fill it
	var targets []Agent
	if err := s.DB.Select(&targets, query); err != nil {
		return nil, fmt.Errorf("GetTargets: %w", err)
	}

	return targets, nil
}

// SetTargets clears all targeted agents then sets the provided agents as targeted
func (s *Storage) SetTargets(agentIDs []string) error {
	// Prevents invalid SQL 'IN ()'
	if len(agentIDs) == 0 {
		return nil
	}

	// Clear targets
	if err := s.ClearTargets(); err != nil {
		return fmt.Errorf("SetTargets: %w", err)
	}

	// Set agents as targeted
	if err := s.AddTargets(agentIDs); err != nil {
		return fmt.Errorf("SetTargets: %w", err)
	}

	// No errors
	return nil
}

func (s *Storage) InsertTask(t *Task) error {
	// Query for the db
	query := `
INSERT INTO tasks (agent_id, type, completed, payload, timeout)
VALUES (?, ?, ?, ?, ?)
`

	// Insert the task into the db
	res, err := s.DB.Exec(query, t.AgentID, t.Type, t.Completed, t.Payload, t.Timeout)
	if err != nil {
		return fmt.Errorf("InsertTask: %w", err)
	}

	// Grab the id from the db and assign it to the struct
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("InsertTask: %w", err)
	}
	t.TaskID = int(id)

	// No errors
	return nil
}

// GetResults returns results for all agents or a specific agent
// An empty agentID will return results of all agents
func (s *Storage) GetResults(agentID string) ([]Result, error) {
	// Initialize variables
	var (
		results []Result
		// Allows for expansion
		args []any
	)
	query := `
SELECT r.result_id, r.agent_id, r.task_id, r.output, r.return_code, t.type, t.payload
FROM results r
JOIN tasks t ON r.task_id = t.task_id
`

	// If agentID is provided, modify the query to reflect that
	if agentID != "" {
		query += ` WHERE agent_id = ?`
		// Make args expand to the agent id
		args = append(args, agentID)
	}

	// args expands to agent id or nothing if no agent id was specified
	if err := s.DB.Select(&results, query, args...); err != nil {
		return nil, fmt.Errorf("GetResults: %w", err)
	}
	return results, nil
}
