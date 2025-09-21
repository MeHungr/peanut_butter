package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

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
