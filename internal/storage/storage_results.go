package storage

import (
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

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

// GetResults returns results for all agents or a specific agent
// An empty agentID will return results of all agents
func (s *Storage) GetResults(filter AgentFilter, limit int) ([]Result, error) {
	// Initialize variables
	var results []Result
	base := `
SELECT r.result_id, r.agent_id, r.task_id, r.output, r.return_code, r.created_at, t.type, t.payload, a.os
FROM results r
JOIN tasks t ON r.task_id = t.task_id
JOIN agents a ON r.agent_id = a.agent_id
	`
	var whereClause []string
	var args []any

	// If the filter is not searching for all agents
	if !filter.All {
		// If filtering by id
		if len(filter.IDs) > 0 {
			clause, idArgs, _ := sqlx.In("r.agent_id IN (?)", filter.IDs)
			whereClause = append(whereClause, clause)
			args = append(args, idArgs...)
		}
		// If filtering by OS
		if len(filter.OSes) > 0 {
			clause, osArgs, _ := sqlx.In("a.os IN (?)", filter.OSes)
			whereClause = append(whereClause, clause)
			args = append(args, osArgs...)
		}
	}

	// Add the query filters to the base query
	query := base
	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
	}

	query += ` ORDER BY r.created_at DESC`

	// If a limit is provided, modify the query to limit results
	if limit > 0 {
		query += ` LIMIT ?`
		// Add limit to args
		args = append(args, limit)
	}

	// args expands to agent id or nothing if no agent id was specified
	if err := s.DB.Select(&results, query, args...); err != nil {
		return nil, fmt.Errorf("GetResults: %w", err)
	}
	return results, nil
}
