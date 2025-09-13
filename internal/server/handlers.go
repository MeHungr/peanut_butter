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
	storageAgent := apiToStorageAgent(agent)
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

	// Checks that the agent id exists
	if _, ok := agents[agent.ID]; !ok {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Updates LastSeen
	now := time.Now()
	agents[agent.ID].LastSeen = &now

	// Look up tasks for this agent
	agentTasks, ok := tasks[agent.ID]
	if !ok || len(agentTasks) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Selects the first uncompleted task
	var agentTask *api.Task
	for _, task := range agentTasks {
		if task.Completed == false {
			agentTask = task
			break
		}
	}
	if agentTask == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Return first task as json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(*agentTask); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	agentTask = nil
}

// ResultHandler allows sending the results from the agent to the server
func ResultHandler(w http.ResponseWriter, r *http.Request) {
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

	// Checks that the agent id exists
	if _, ok := agents[result.AgentID]; !ok {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Updates the agent's last seen time
	now := time.Now().UTC()
	agents[result.AgentID].LastSeen = &now

	// Updates the task to be completed
	var taskToUpdate *api.Task
	agentTasks := tasks[result.AgentID]
	for _, task := range agentTasks {
		if task.TaskID == result.TaskID {
			taskToUpdate = task
			break
		}
	}
	if taskToUpdate == nil {
		http.Error(w, "Task does not exist", http.StatusBadRequest)
		return
	}
	taskToUpdate.Completed = true

	// Prints to the console the task being completed
	fmt.Printf("\nAgent %s completed task: \"%s\"\n\n Return Code: %d, Output:\n%s", agents[result.AgentID].ID, result.Payload, result.ReturnCode, result.Output)

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
func GetAgentsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensures GET method is used
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Initializes a slice of agent copies and adds agents from the agents map
	var agentList api.GetAgentsResponse
	for _, agent := range agents {
		agentList.AgentIDs = append(agentList.AgentIDs, agent)
	}

	// Encodes JSON and sends message
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agentList); err != nil {
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
