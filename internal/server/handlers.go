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

	log.Printf("Registering agent with ID: %q", agent.ID)

	// Convert the api.Agent to a storage.Agent
	storageAgent := apiToStorageAgent(&agent)
	// Attempt to register the agent with the db
	if err := srv.storage.RegisterAgent(storageAgent); err != nil {
		log.Printf("RegisterAgent failed for %s: %v", agent.ID, err)
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
		log.Printf("Failed to update last_seen for %s: %v", agent.ID, err)
	}

	// Retrieves the next task from the db
	task, err := srv.storage.GetNextTask(agent.ID)
	if err != nil {
		log.Printf("GetNextTask failed for %s: %v", agent.ID, err)
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
		log.Printf("Failed to update last_seen for %s: %v", result.AgentID, err)
	}

	// Convert api.Result to a storage.Result
	storageResult := apiToStorageResult(&result)
	// Insert the result into the db
	if err := srv.storage.InsertResult(storageResult); err != nil {
		log.Printf("Failed to insert result for task %d: %v", result.TaskID, err)
		http.Error(w, "Failed to store result", http.StatusInternalServerError)
		return
	}

	// Mark the corresponding task as completed
	if err := srv.storage.MarkTaskCompleted(result.TaskID); err != nil {
		log.Printf("Failed to mark task as completed for task %d: %v", result.TaskID, err)
		http.Error(w, "Failed to mark task as completed", http.StatusInternalServerError)
		return
	}

	// Prints to the console the task being completed
	log.Printf(`
[agent=%s task=%d] payload="%s" return_code=%d
Output:
%s
`,
		result.AgentID, result.TaskID, result.Payload, result.ReturnCode, result.Output)

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

	agents, err := srv.storage.GetAllAgents()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	// Initializes a slice of agent copies and adds agents from the agents map
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
func AddTargetsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Mark the requested agents as targeted
	for _, id := range reqBody.AgentIDs {
		if agent, ok := agents[id]; ok {
			agent.Targeted = true
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

// GetTargetsHandler sends a list of current targets
func GetTargetsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Iterate through agents and append targeted agents to targeted
	var targeted []*api.Agent
	for _, agent := range agents {
		if agent.Targeted {
			targeted = append(targeted, agent)
		}
	}

	resp := api.GetTargetsResponse{
		Agents: targeted,
		Count:  len(targeted),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UntargetHandler allows agents to be untargeted
func UntargetHandler(w http.ResponseWriter, r *http.Request) {
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

	// Iterate through the provided agents and set Targeted to false
	for _, id := range reqBody.AgentIDs {
		if agent, ok := agents[id]; ok {
			agent.Targeted = false
		}
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
func ClearTargetsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set all agents.Targeted to false
	for _, agent := range agents {
		agent.Targeted = false
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
func SetTargetsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Set all agents.Targeted to false
	for _, agent := range agents {
		agent.Targeted = false
	}

	// Set only the given agents
	for _, id := range reqBody.AgentIDs {
		if agent, ok := agents[id]; ok {
			agent.Targeted = true
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

// EnqueueHandler allows enqueueing of tasks
func EnqueueHandler(w http.ResponseWriter, r *http.Request) {
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

	task := &api.Task{}
	// Decode the JSON into a task, error if fail
	if err := json.NewDecoder(r.Body).Decode(task); err != nil {
		http.Error(w, "Invalid task", http.StatusUnsupportedMediaType)
		return
	}

	// Ensures the request included an agent_id
	if task.AgentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	// Ensure the task is valid
	if strings.TrimSpace(task.Payload) == "" {
		http.Error(w, "Task payload required", http.StatusBadRequest)
		return
	}

	// Make sure the agent exists
	if _, ok := agents[task.AgentID]; !ok {
		http.Error(w, "Agent not found", http.StatusBadRequest)
		return
	}

	// Timestamp the task and assign an id
	now := time.Now()
	task.Timestamp = &now

	// Add the task to the agent's queue
	tasks[task.AgentID] = append(tasks[task.AgentID], task)

	// Set content type and send 200 status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Message JSON format
	msg := api.Message{
		Message: "Task enqueued",
	}

	// Marshal and send JSON
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	// Log to server
	fmt.Printf("Enqueued task for agent %s: %s\n", task.AgentID, task.Payload)
}
