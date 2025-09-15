package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// RegisterHandler handles the registration of an agent to the server
// The /register endpoint expects an agent_id in a POST request
func (srv *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("Registering agent with ID: %q\n", agent.ID)

	// Convert the api.Agent to a storage.Agent
	storageAgent := apiToStorageAgent(&agent)
	// Attempt to register the agent with the db
	if err := srv.storage.RegisterAgent(storageAgent); err != nil {
		log.Printf("RegisterAgent failed for %s: %v\n", agent.ID, err)
		http.Error(w, "Failed to register agent", http.StatusInternalServerError)
		return
	}

	// Sends back a registered message
	msg := api.Message{
		Message: "Registered",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	log.Printf("Agent: %s has registered\n", agent.ID)
}
