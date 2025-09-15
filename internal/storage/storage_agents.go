package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

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
