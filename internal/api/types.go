// Defines the API and structure of communication between agents and the server
package api

import (
	"time"
)

// Agent represents a single registered agent in the C2
type Agent struct {
	ID               string        `json:"agent_id"`
	OS               string        `json:"os,omitempty"`
	Arch             string        `json:"arch,omitempty"`
	AgentIP          string        `json:"agent_ip,omitempty"`
	ServerIP         string        `json:"server_ip,omitempty"`
	ServerPort       int           `json:"server_port,omitempty"`
	CallbackInterval time.Duration `json:"callback_interval,omitempty"`
	Hostname         string        `json:"hostname,omitempty"`
	Status           string        `json:"status,omitempty"`
	LastSeen         *time.Time    `json:"last_seen,omitempty"`
}

// Task represents a task for an agent to complete, served by the server
type Task struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Completed bool           `json:"completed,omitempty"`
	Payload   string         `json:"payload,omitempty"`
	Timeout   *time.Duration `json:"timeout,omitempty"`
	Timestamp *time.Time     `json:"timestamp,omitempty"`
}

// Result represents the result of a completed task, returned by an agent
type Result struct {
	TaskID     string `json:"task_id"`
	AgentID    string `json:"agent_id"`
	Output     string `json:"output"`
	ReturnCode int    `json:"return_code"`
}
