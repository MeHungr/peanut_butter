// Contains the structs used for interacting with the database
package storage

import (
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// Agent represents a single registered agent in the C2
type Agent struct {
	ID               string        `db:"agent_id"`
	OS               string        `db:"os"`
	Arch             string        `db:"arch"`
	Targeted         bool          `db:"targeted"`
	AgentIP          string        `db:"agent_ip"`
	ServerIP         string        `db:"server_ip"`
	ServerPort       int           `db:"server_port"`
	CallbackInterval time.Duration `db:"callback_interval"`
	Hostname         string        `db:"hostname"`
	LastSeen         *time.Time    `db:"last_seen"`
}

// Task represents a task for an agent to complete, served by the server
type Task struct {
	TaskID    int            `db:"task_id"`
	AgentID   string         `db:"agent_id"`
	Type      api.TaskType   `db:"type"`
	Completed bool           `db:"completed"`
	Payload   string         `db:"payload"`
	Timeout   *time.Duration `db:"timeout"`
	Timestamp *time.Time     `db:"timestamp"`
}

// Result represents the result of a completed task, returned by an agent
type Result struct {
	ResultID   int    `db:"result_id"`
	TaskID     int    `db:"task_id"`
	AgentID    string `db:"agent_id"`
	Output     string `db:"output"`
	ReturnCode int    `db:"return_code"`
	Type       string `db:"type"`
	Payload    string `db:"payload"`
}
