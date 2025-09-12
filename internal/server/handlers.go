// Includes HTTP handlers for use in the server package
package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/google/uuid"
)

func requireLocalhost(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || (host != "127.0.0.1" && host != "::1") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

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
		if task.TaskID == result.TaskID {
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
	fmt.Printf("\nAgent %s completed task: '%s'\n\n Return Code: %d, Output:\n%s", agents[result.AgentID].ID, result.Payload, result.ReturnCode, result.Output)

	w.WriteHeader(http.StatusOK)
}

// GetAgentsHandler returns the list of connected agents
func GetAgentsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensures GET method is used
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Initializes a slice of agent copies and adds agents from the agents map
	var agentList []api.Agent
	for _, agent := range agents {
		agentList = append(agentList, *agent)
	}

	// Encodes JSON and sends message
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agentList); err != nil {
		http.Error(w, "Failed to encode agents", http.StatusInternalServerError)
		return
	}
}

// EnqueueHandler allows enqueueing of tasks
func EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	// Only accepts POST requests
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

	task := &api.Task{}
	// Decode the JSON into a task, error if fail
	if err := json.NewDecoder(r.Body).Decode(task); err != nil {
		http.Error(w, "Invalid task", http.StatusUnsupportedMediaType)
		return
	}

	// Ensures the request included an agent_id
	if task.AgentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	// Ensure the task is valid
	if strings.TrimSpace(task.Payload) == "" {
		http.Error(w, "Task payload required", http.StatusBadRequest)
		return
	}

	// Make sure the agent exists
	if _, ok := agents[task.AgentID]; !ok {
		http.Error(w, "Agent not found", http.StatusBadRequest)
		return
	}

	// Timestamp the task and assign an id
	now := time.Now()
	task.Timestamp = &now
	task.TaskID = uuid.New().String()

	// Add the task to the agent's queue
	tasks[task.AgentID] = append(tasks[task.AgentID], task)

	// Set content type and send 200 status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Message JSON format
	msg := map[string]string{
		"message": "Task enqueued",
	}

	// Marshal and send JSON
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	// Log to server
	fmt.Printf("Enqueued task for agent %s: %s\n", task.AgentID, task.Payload)
}
