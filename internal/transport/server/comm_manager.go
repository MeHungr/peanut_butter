package server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/conversion"
	"github.com/MeHungr/peanut-butter/internal/pberrors"
	"github.com/MeHungr/peanut-butter/internal/storage"
	"github.com/MeHungr/peanut-butter/internal/transport"
)

type CommManager struct {
	Transports map[transport.TransportString]Transport
	Agents     map[string]Transport // agentID -> Transport
	Storage    *storage.Storage     // db used by server
}

func (cm *CommManager) RegisterAgent(agent *api.Agent, t Transport) error {
	// Validates that the agent ID is non-empty
	if agent.AgentID == "" {
		return pberrors.ErrInvalidAgentID
	}

	// Updates the agent's last seen time
	now := time.Now().UTC()
	agent.LastSeen = &now

	log.Printf("Registering agent with ID: %q\n", agent.AgentID)

	// Convert the api.Agent to a storage.Agent
	storageAgent := conversion.ApiToStorageAgent(agent)
	// Attempt to register the agent with the db
	if err := cm.Storage.RegisterAgent(storageAgent); err != nil {
		return fmt.Errorf("RegisterAgent failed for %s: %w\n", agent.AgentID, err)
	}

	cm.Agents[agent.AgentID] = t

	return nil
}

func (cm *CommManager) SendTask(agent *api.Agent) (*api.Task, error) {
	// Validates that the agent ID is non-empty
	if agent.AgentID == "" {
		return nil, pberrors.ErrInvalidAgentID
	}

	// Retrives the agent from the db
	storageAgent, err := cm.Storage.GetAgentByID(agent.AgentID)
	if err != nil {
		return nil, fmt.Errorf("Sendtask failed for %s: %w\n", agent.AgentID, err)
	}

	// If storageAgent is nil, no agent with the id exists
	if storageAgent == nil {
		return nil, fmt.Errorf("SendTask failed for %s: %w\n", agent.AgentID, err)
	}

	// Update agent's last seen time to now
	now := time.Now().UTC()
	if err := cm.Storage.UpdateLastSeen(agent.AgentID, now); err != nil {
		return nil, fmt.Errorf("Failed to update last_seen for %s: %w\n", agent.AgentID, err)
	}

	// Retrieves the next task from the db
	task, err := cm.Storage.GetNextTask(agent.AgentID)
	if err != nil {
		return nil, fmt.Errorf("GetNextTask failed for %s: %w\n", agent.AgentID, err)
	}

	// If task is nil, no tasks were found
	if task == nil {
		return nil, nil
	}

	// Convert the task to be used with the api
	apiTask := conversion.StorageToAPITask(task)

	transport, ok := cm.Agents[agent.AgentID]
	if !ok {
		return nil, fmt.Errorf("no transport found for agent %s", agent.AgentID)
	}

	return transport.SendTask(apiTask)
}

func (cm *CommManager) HandleResult(result *api.Result) error {
	// Validates that the agent ID is non-empty
	if result.AgentID == "" {
		return pberrors.ErrInvalidAgentID
	}

	// Validates that the task ID is non-empty
	if result.TaskID == 0 {
		return pberrors.ErrInvalidTaskID
	}

	// Retrieve the agent and handle errors
	storageAgent, err := cm.Storage.GetAgentByID(result.AgentID)
	if err != nil {
		return fmt.Errorf("HandleResult failed for result %q: %w\n", result.ResultID, err)

	}
	if storageAgent == nil {
		return fmt.Errorf("Agent %q does not exist: %w\n", result.AgentID, err)
	}

	// Updates the agent's last seen time
	now := time.Now().UTC()
	if err := cm.Storage.UpdateLastSeen(result.AgentID, now); err != nil {
		log.Printf("Failed to update last_seen for %q: %v\n", result.AgentID, err)
	}

	// Convert api.Result to a storage.Result
	storageResult := conversion.ApiToStorageResult(result)
	// Insert the result into the db
	if err := cm.Storage.InsertResult(storageResult); err != nil {
		log.Printf("Failed to insert result for task %d: %v\n", result.TaskID, err)
		return fmt.Errorf("Failed to store result %q: %w\n", result.ResultID, err)
	}

	// Mark the corresponding task as completed
	if err := cm.Storage.MarkTaskCompleted(result.TaskID); err != nil {
		log.Printf("Failed to mark task as completed for task %d: %v\n", result.TaskID, err)
		return fmt.Errorf("Failed to mark task %q as completed: %w\n", result.TaskID, err)
	}

	// Truncate output for logs
	out := strings.SplitN(result.Output, "\n", 2)[0] // first line only
	if len(out) > 80 {
		out = out[:77] + "..."
	}
	// Prints to the console the task being completed
	log.Printf(`[agent=%s task=%d rc=%d] payload=%q output=%q
`,
		result.AgentID, result.TaskID, result.ReturnCode, result.Payload, out)

	return nil
}
