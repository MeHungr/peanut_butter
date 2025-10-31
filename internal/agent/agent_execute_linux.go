// Contains the logic for executing commands in linux
//go:build linux

package agent

import (
	"context"
	"os/exec"
	"strconv"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// executeCommand takes a task and returns the output and returnCode of the command after executing
func executeCommand(task *api.Task) (output string, returnCode string) {
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
		cmd = exec.CommandContext(ctx, "bash", "-c", task.Payload)
	} else { // If the task has no timeout, just execute it
		cmd = exec.Command("bash", "-c", task.Payload)
	}

	// Runs the command and captures stdout + stderr
	out, err := cmd.CombinedOutput()
	output = string(out)

	// Check if the context timed out
	if task.Timeout != nil && ctx.Err() == context.DeadlineExceeded {
		returnCode = CmdTimeout
		return
	}

	// If the command failed for another reason
	if err != nil {
		if cmd.ProcessState != nil {
			returnCode = strconv.Itoa(cmd.ProcessState.ExitCode())
		} else {
			returnCode = FailedToStart // Failed to start process
		}
		return
	}

	return // Returns the named values in the function signature
}
