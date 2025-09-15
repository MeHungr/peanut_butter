package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// ResultHandler allows sending the results from the agent to the server
func (srv *Server) ResultHandler(w http.ResponseWriter, r *http.Request) {
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
	if result.TaskID == 0 {
		http.Error(w, "No task ID", http.StatusBadRequest)
		return
	}

	// Retrieve the agent and handle errors
	storageAgent, err := srv.storage.GetAgentByID(result.AgentID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if storageAgent == nil {
		http.Error(w, "Agent does not exist", http.StatusBadRequest)
		return
	}

	// Updates the agent's last seen time
	now := time.Now().UTC()
	if err := srv.storage.UpdateLastSeen(result.AgentID, now); err != nil {
		log.Printf("Failed to update last_seen for %s: %v\n", result.AgentID, err)
	}

	// Convert api.Result to a storage.Result
	storageResult := apiToStorageResult(&result)
	// Insert the result into the db
	if err := srv.storage.InsertResult(storageResult); err != nil {
		log.Printf("Failed to insert result for task %d: %v\n", result.TaskID, err)
		http.Error(w, "Failed to store result", http.StatusInternalServerError)
		return
	}

	// Mark the corresponding task as completed
	if err := srv.storage.MarkTaskCompleted(result.TaskID); err != nil {
		log.Printf("Failed to mark task as completed for task %d: %v\n", result.TaskID, err)
		http.Error(w, "Failed to mark task as completed", http.StatusInternalServerError)
		return
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

	msg := api.Message{
		Message: "Result received",
	}

	// Send back response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&msg); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetResultsHandler retrieves results from the db and responds with a slice of results
func (srv *Server) GetResultsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// optional ?agent_id= query parameter
	agentID := r.URL.Query().Get("agent_id")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Invalid query parameter 'limit': %w", http.StatusBadRequest)
		return
	}

	// Get results from db
	results, err := srv.storage.GetResults(agentID, limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Generate response
	resp := api.GetResultsResponse{
		Results: storagetoAPIResults(results),
	}

	// Marshal response and send
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
