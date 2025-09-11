// Includes HTTP handlers for use in the server package
package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// RegisterHandler handles the registration of an agent to the server
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Check that the HTTP method is POST. This is the only allowed method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	
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
	agents[agent.ID] = agent

	// Sends back a registered message
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Registered\n"))
}

// TaskHandler handles the distribution of tasks to agents
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

	// Look up tasks for this agent
	agentTasks, ok := tasks[agent.ID]
	if !ok || len(agentTasks) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Return first task as json
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(agentTasks[0])
	// Throws an error if encoding fails
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	// Pops the first task from the map
	tasks[agent.ID] = agentTasks[1:]
}

// // EnqueueHandler adds a task for a specific agent
// func EnqueueHandler(w http.ResponseWriter, r *http.Request) {
//     if r.Method != http.MethodPost {
//         http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//         return
//     }
//     defer r.Body.Close()
//
//     if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
//         http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
//         return
//     }
//
//     // Envelope for input
//     var input struct {
//         AgentID string    `json:"agent_id"`
//         Task    api.Task  `json:"task"`
//     }
//
//     if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
//         http.Error(w, "Invalid JSON", http.StatusBadRequest)
//         return
//     }
//
//     if input.AgentID == "" {
//         http.Error(w, "agent_id required", http.StatusBadRequest)
//         return
//     }
//
//     // Append task to that agent’s queue
//     tasks[input.AgentID] = append(tasks[input.AgentID], input.Task)
//
//     w.Header().Set("Content-Type", "text/plain; charset=utf-8")
//     w.WriteHeader(http.StatusOK)
//     w.Write([]byte("Task enqueued\n"))
// }
