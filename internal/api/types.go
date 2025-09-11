package api

import (
	"time"
)

type Agent struct {
	ID       string    `json:"agent_id"`
	OS       string    `json:"os,omitempty"`
	Arch     string    `json:"arch,omitempty"`
	Hostname string    `json:"hostname,omitempty"`
	Status   string      `json:"status,omitempty"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

type Task struct {
	ID        string    `json:"task_id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type Result struct {
	TaskID  string `json:"task_id"`
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
	Output  string `json:"output"`
}
