// Includes HTTP handlers for use in the server package
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// RegisterHandler handles the registration of an agent to the server
// The /register endpoint expects an agent_id in a POST request
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Check that the HTTP method is POST. This is the only allowed method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	// Ensures json content type
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	// Declares a new Agent for decoding
	var agent api.Agent

	// Decodes the JSON of the incoming request into the agent variable and checks for errors
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validates that the agent ID is non-empty
	if agent.ID == "" {
		http.Error(w, "No agent ID", http.StatusBadRequest)
		return
	}

	// Updates the agent's last seen time
	now := time.Now().UTC()
	agent.LastSeen = &now

	// Maps the agent's id to the agent itself
	agents[agent.ID] = &agent

	// Sends back a registered message
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Registered\n"))
}

// TaskHandler handles the distribution of tasks to agents
// The /task endpoint expects an agent_id in a POST request
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	// Check that the HTTP method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var agent api.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validates that the agent ID is non-empty
	if agent.ID == "" {
		http.Error(w, "No agent ID", http.StatusBadRequest)
		return
	}

	// Checks that the agent id exists
	if _, ok := agents[agent.ID]; !ok {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Updates LastSeen
	now := time.Now()
	agents[agent.ID].LastSeen = &now

	// Look up tasks for this agent
	agentTasks, ok := tasks[agent.ID]
	if !ok || len(agentTasks) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Selects the first uncompleted task
	var agentTask *api.Task
	for _, task := range agentTasks {
		if task.Completed == false {
			agentTask = task
			break
		}
	}
	if agentTask == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Return first task as json
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(*agentTask)
	// Throws an error if encoding fails
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	agentTask = nil
}

// ResultHandler allows sending the results from the agent to the server
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	// Check that the HTTP method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	// Ensures json content type
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	// Declares an empty Result and decodes the POST body into it
	var result api.Result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validates that the agent ID is non-empty
	if result.AgentID == "" {
		http.Error(w, "No agent ID", http.StatusBadRequest)
		return
	}

	// Validates that the task ID is non-empty
	if result.TaskID == "" {
		http.Error(w, "No task ID", http.StatusBadRequest)
		return
	}

	// Checks that the agent id exists
	if _, ok := agents[result.AgentID]; !ok {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Updates the agent's last seen time
	now := time.Now().UTC()
	agents[result.AgentID].LastSeen = &now

	// Updates the task to be completed
	var taskToUpdate *api.Task
	agentTasks := tasks[result.AgentID]
	for _, task := range agentTasks {
		if task.ID == result.TaskID {
			taskToUpdate = task
			break
		}
	}
	if taskToUpdate == nil {
		http.Error(w, "Task does not exist", http.StatusBadRequest)
		return
	}
	taskToUpdate.Completed = true

	// Prints to the console the task being completed
	fmt.Printf("Agent %s completed task: %s\n", agents[result.AgentID].ID, result.TaskID)

	w.WriteHeader(http.StatusOK)
}

// EnqueueHandler adds a task for a specific agent
func EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	// Envelope for input
	var input struct {
		AgentID string   `json:"agent_id"`
		Task    api.Task `json:"task"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.AgentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	// Append task to that agent’s queue
	tasks[input.AgentID] = append(tasks[input.AgentID], &input.Task)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task enqueued\n"))
}
