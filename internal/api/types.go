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
	Targeted         bool          `json:"targeted,omitempty"`
	AgentIP          string        `json:"agent_ip,omitempty"`
	ServerIP         string        `json:"server_ip,omitempty"`
	ServerPort       int           `json:"server_port,omitempty"`
	CallbackInterval time.Duration `json:"callback_interval,omitempty"`
	Hostname         string        `json:"hostname,omitempty"`
	LastSeen         *time.Time    `json:"last_seen,omitempty"`
}

// Defines the TaskType type (see below)
type TaskType string

// Defines the possible types of Tasks for Agents to use
const (
	Command TaskType = "command"
)

// Task represents a task for an agent to complete, served by the server
type Task struct {
	TaskID    int            `json:"task_id"`
	AgentID   string         `json:"agent_id"`
	Type      TaskType       `json:"type"`
	Completed bool           `json:"completed,omitempty"`
	Payload   string         `json:"payload,omitempty"`
	Timeout   *time.Duration `json:"timeout,omitempty"`
	Timestamp *time.Time     `json:"timestamp,omitempty"`
}

// Result represents the result of a completed task, returned by an agent
type Result struct {
	ResultID   int `json:"result_id"`
	Task       `json:"task"`
	Output     string `json:"output"`
	ReturnCode int    `json:"return_code"`
}

// GetAgentsResponse represents the response from the server on the /get-agents endpoint
type GetAgentsResponse struct {
	Agents []*Agent `json:"agents"`
}

// AddTargetsRequest represents the request the agent sends on the /add-targets endpoint
type TargetsRequest struct {
	AgentIDs []string `json:"agents"`
}

// GetTargetsResponse represents the response from the server on the /get-targets endpoint
type GetTargetsResponse struct {
	Agents []*Agent `json:"agents"`
	Count  int      `json:"count"`
}

// EnqueueRequest represents a request to enqueue a task on the /enqueue endpoint
type EnqueueRequest struct {
	Type    TaskType `json:"type"`
	Payload string   `json:"payload"`
	Timeout int      `json:"timeout,omitempty"` // seconds
}

// Message represents a json message sent by the server
type Message struct {
	Message string `json:"message"`
}
