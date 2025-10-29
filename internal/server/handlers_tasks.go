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
	targets, err := srv.storage.GetTargets(storage.AgentFilter{All: true})
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
			AgentID:   a.AgentID,
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
			log.Printf("Failed to insert task for agent %s: %v", a.AgentID, err)
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
