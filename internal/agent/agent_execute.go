package agent

import (
	"fmt"
	"strings"

	"github.com/MeHungr/peanut-butter/internal/api"
)

const (
	CmdTimeout    = "command timed out"
	FailedToStart = "command failed to start"
)

// ExecuteTask executes the task retrieved by GetTask
func (a *Agent) ExecuteTask(task *api.Task) (*api.Result, error) {
	if strings.TrimSpace(task.Payload) == "" {
		return &api.Result{Output: "No task payload"}, nil
	}

	// Declares the result and its agent id
	result := &api.Result{
		Task: *task,
	}

	switch task.Type {
	case api.Command:
		result.Output, result.ReturnCode = executeCommand(task)
	default:
		return result, fmt.Errorf("Undefined task type in JSON: %s", task.Type)
	}

	return result, nil
}
