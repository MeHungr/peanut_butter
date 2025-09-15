package server

import (
	"encoding/json"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// GetTargetsHandler sends a list of current targets
func (srv *Server) GetTargetsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query database for targets
	targets, err := srv.storage.GetTargets()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Convert storage.Agents to api.Agents and add to response body
	var resp api.GetTargetsResponse
	for _, t := range targets {
		resp.Agents = append(resp.Agents, storageToAPIAgent(&t))
	}
	resp.Count = len(resp.Agents)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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

	// Adds the requested agents as targets
	if err := srv.storage.AddTargets(reqBody.AgentIDs); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
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
	if err := srv.storage.Untarget(reqBody.AgentIDs); err != nil {
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

	// Clear then set targets
	if err := srv.storage.SetTargets(reqBody.AgentIDs); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
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
