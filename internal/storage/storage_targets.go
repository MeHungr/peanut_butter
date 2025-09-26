package storage

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// GetTargets handles retrieving of agents given a filter (or none)
func (s *Storage) GetTargets(filter AgentFilter) ([]Agent, error) {
	base := `SELECT * FROM agents WHERE targeted = 1`
	var whereClause []string
	var args []any

	// If the filter is not searching for all agents
	if !filter.All {
		// If filtering by id
		if len(filter.IDs) > 0 {
			clause, idArgs, _ := sqlx.In("id IN (?)", filter.IDs)
			whereClause = append(whereClause, clause)
			args = append(args, idArgs...)
		}
		// If filtering by OS
		if len(filter.OSes) > 0 {
			clause, osArgs, _ := sqlx.In("os IN (?)", filter.OSes)
			whereClause = append(whereClause, clause)
			args = append(args, osArgs...)
		}
		// If filtering by status
		if len(filter.Statuses) > 0 {
			clause, statusArgs, _ := sqlx.In("status IN (?)", filter.Statuses)
			whereClause = append(whereClause, clause)
			args = append(args, statusArgs...)
		}
	}

	// Add the query filters to the base query
	query := base
	if len(whereClause) > 0 {
		query += " AND " + strings.Join(whereClause, " AND ")
	}

	// Ensure the query works with the correct database syntax
	query = s.DB.Rebind(query)

	var targets []Agent
	if err := s.DB.Select(&targets, query, args...); err != nil {
		return nil, fmt.Errorf("GetAgents: %w", err)
	}

	return targets, nil
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

// TargetAll sets all agents to targeted
func (s *Storage) TargetAll() error {
	// Query for the db
	query := `UPDATE agents SET targeted = 1`

	// Make all agents targeted
	if _, err := s.DB.Exec(query); err != nil {
		return fmt.Errorf("TargetAll: %w", err)
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
