package server

import (
	"encoding/json"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// GetAgentsHandler returns the list of connected agents
func (srv *Server) GetAgentsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensures GET method is used
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Queries database for all agents
	agents, err := srv.storage.GetAllAgents()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Converts storage.Agents to api.Agents and adds them to response body
	var resp api.GetAgentsResponse
	var count int
	for _, a := range agents {
		resp.Agents = append(resp.Agents, storageToAPIAgent(&a))
		count++
	}
	resp.Count = count

	// Encodes JSON and sends message
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
