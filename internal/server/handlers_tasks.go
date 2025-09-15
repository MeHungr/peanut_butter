package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

// TaskHandler handles the distribution of tasks to agents
// The /task endpoint expects an agent_id in a POST request
func (srv *Server) TaskHandler(w http.ResponseWriter, r *http.Request) {
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

	// Retrives the agent from the db
	storageAgent, err := srv.storage.GetAgentByID(agent.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If storageAgent is nil, no agent with the id exists
	if storageAgent == nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Update agent's last seen time to now
	now := time.Now().UTC()
	if err := srv.storage.UpdateLastSeen(agent.ID, now); err != nil {
		log.Printf("Failed to update last_seen for %s: %v\n", agent.ID, err)
	}

	// Retrieves the next task from the db
	task, err := srv.storage.GetNextTask(agent.ID)
	if err != nil {
		log.Printf("GetNextTask failed for %s: %v\n", agent.ID, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If task is nil, no tasks were found
	if task == nil {
		w.WriteHeader(http.StatusNoContent) // No tasks for this agent
		return
	}

	// Convert the task to be used with the api
	apiTask := storageToAPITask(task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiTask); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

// EnqueueHandler allows enqueueing of tasks
func (srv *Server) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
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

	var req api.EnqueueRequest
	// Decode the JSON into an api.EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid task", http.StatusUnsupportedMediaType)
		return
	}

	// Get all targeted agents
	targets, err := srv.storage.GetTargets()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Enqueue a task for each targeted agent
	count := 0
	for _, a := range targets {
		// Timestamp for the task
		now := time.Now().UTC()

		// Create the task
		task := storage.Task{
			AgentID:   a.ID,
			Type:      req.Type,
			Completed: false,
			Payload:   req.Payload,
			Timestamp: &now,
		}
		// Only include timeout if > 0
		if req.Timeout > 0 {
			dur := time.Duration(req.Timeout) * time.Second
			task.Timeout = &dur
		}
		// Attempt to insert the task into the db
		if err := srv.storage.InsertTask(&task); err != nil {
			log.Printf("Failed to insert task for agent %s: %v", a.ID, err)
			continue
		}
		count++
	}

	// Message JSON format
	msg := api.Message{
		Message: fmt.Sprintf("Enqueued %d tasks", count),
	}

	// Set content type and send 200 status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Marshal and send JSON
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Truncate payload if too long
	out := req.Payload
	if len(out) > 80 {
		out = out[:77] + "..."
	}

	// Log to server
	log.Printf("[enqueue count=%d type=%s] payload=%q", count, req.Type, out)
}
