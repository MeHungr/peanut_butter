package server

import (
	"encoding/json"
	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/storage"
	"net/http"
)

// GetTargetsHandler returns the list of targeted agents
func (srv *Server) GetTargetsHandler(w http.ResponseWriter, r *http.Request) {
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
	agents, err := srv.storage.GetTargets(filter)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Converts storage.Agents to api.Agents and adds them to response body
	var resp api.GetAgentsResponse
	resp.Count = len(agents)
	for _, a := range agents {
		resp.Agents = append(resp.Agents, storageToAPIAgent(&a))
	}

	// Encodes JSON and sends message
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// AddTargetsHandler allows the cli to add targets to task enqueueing
func (srv *Server) AddTargetsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	// Decode JSON into slice of agent ids in TargetsRequest
	var reqBody api.TargetsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// If TargetAll is true, set all targets, else, add specified targets
	switch reqBody.All {
	case true:
		// If TargetAll is set to true, target all and return
		if err := srv.storage.TargetAll(); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	case false:
		// Adds the requested agents as targets
		if err := srv.storage.AddTargets(apiToStorageFilter(reqBody.AgentFilter)); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	// Response message
	msg := api.Message{
		Message: "Targets added",
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UntargetHandler allows agents to be untargeted
func (srv *Server) UntargetHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	// Decode the json into reqBody
	var reqBody api.TargetsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Untarget the provided agents
	if err := srv.storage.Untarget(apiToStorageFilter(reqBody.AgentFilter)); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send a success message
	msg := api.Message{
		Message: "Targets removed",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ClearTargetsHandler allows clearing of the target list
func (srv *Server) ClearTargetsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear targets
	if err := srv.storage.ClearTargets(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	msg := api.Message{
		Message: "All targets cleared",
	}
	// Send back response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SetTargetsHandler clears the target list then sets the targets to those provided
func (srv *Server) SetTargetsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow PUT
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	// Decode the request body
	var reqBody api.TargetsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// If TargetAll is true, set all targets, else, set specified targets
	switch reqBody.All {
	case true:
		// If TargetAll is set to true, target all and return
		if err := srv.storage.TargetAll(); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	case false:
		// Clear all, then set specified targets
		if err := srv.storage.SetTargets(apiToStorageFilter(reqBody.AgentFilter)); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	msg := api.Message{
		Message: "Targets set",
	}
	// Send back response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
