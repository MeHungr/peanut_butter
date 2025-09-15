package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

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
