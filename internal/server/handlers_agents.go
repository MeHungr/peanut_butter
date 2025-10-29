package server

import (
	"encoding/json"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/conversion"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

// GetAgentsHandler returns the list of connected agents
func (srv *Server) GetAgentsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensures GET method is used
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filter := storage.AgentFilter{}

	// ?all = true returns all agents
	if query.Get("all") == "true" {
		filter.All = true
	}

	// ?id=123&id=456
	if ids, ok := query["id"]; ok {
		filter.IDs = ids
	}

	// ?os=linux&os=windows
	if oses, ok := query["os"]; ok {
		filter.OSes = oses
	}

	// ?status=active&status=inactive
	if statuses, ok := query["status"]; ok {
		filter.Statuses = statuses
	}

	// Query database with filter
	agents, err := srv.storage.GetAgents(filter)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Converts storage.Agents to api.Agents and adds them to response body
	var resp api.GetAgentsResponse
	resp.Count = len(agents)
	for _, a := range agents {
		resp.Agents = append(resp.Agents, conversion.StorageToAPIAgent(&a))
	}

	// Encodes JSON and sends message
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
