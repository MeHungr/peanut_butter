// Includes HTTP handlers for use in the server package
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/storage"
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
	log.Printf(`[agent=%s task=%d rc=%d] payload="%q" output=%q
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
