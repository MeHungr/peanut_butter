package storage

import (
	"fmt"
	"log"
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
func (s *Storage) GetResults(agentID string, limit int) ([]Result, error) {
	// Initialize variables
	var (
		results []Result
		// Allows for expansion
		args []any
	)
	query := `
SELECT r.result_id, r.agent_id, r.task_id, r.output, r.return_code, r.created_at, t.type, t.payload
FROM results r
JOIN tasks t ON r.task_id = t.task_id
`

	// If agentID is provided, modify the query to reflect that
	if agentID != "" {
		query += ` WHERE r.agent_id = ?`
		// Make args expand to the agent id
		args = append(args, agentID)
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
