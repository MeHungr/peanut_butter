package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/pberrors"
)

type HTTPTransport struct {
	Comm *CommManager
	port int
}

func (h *HTTPTransport) Start() error {
	// Defines the /register path and uses RegisterHandler to handle data
	http.HandleFunc("/register", h.RegisterHandler)
	// Defines the /task path and uses TaskHandler to handle data
	http.HandleFunc("/task", h.TaskHandler)
	// Defines the /result path and uses ResultHandler to handle data
	http.HandleFunc("/result", h.ResultHandler)

	return nil
}

// HandleResult implements ServerTransport.
func (h *HTTPTransport) HandleResult(r *api.Result) error {
	return h.Comm.HandleResult(r)
}

// RegisterAgent implements ServerTransport.
func (h *HTTPTransport) RegisterAgent(a *api.Agent) error {
	return h.Comm.RegisterAgent(a, h)
}

// SendTask implements ServerTransport.
func (h *HTTPTransport) SendTask(t *api.Task) (*api.Task, error) {
	return t, nil
}

// RegisterHandler handles the registration of an agent to the server
// The /register endpoint expects an RegisterRequest in a POST body
func (h *HTTPTransport) RegisterHandler(w http.ResponseWriter, r *http.Request) {
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

	// Declares a new request for decoding
	var registerReq api.RegisterRequest

	// Decodes the JSON of the incoming request into the agent variable and checks for errors
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// If no agent is sent, return bad request error
	if registerReq.Agent == nil {
		http.Error(w, "Missing agent", http.StatusBadRequest)
		return
	}

	// Defines the agent given by the request
	agent := registerReq.Agent

	if err := h.RegisterAgent(agent); err != nil {
		var status int
		switch {
		case errors.Is(err, pberrors.ErrInvalidAgentID):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), status)
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
	log.Printf("Agent: %s has registered\n", agent.AgentID)
}

// TaskHandler handles the distribution of tasks to agents
// The /task endpoint expects an agent_id in a POST request
func (h *HTTPTransport) TaskHandler(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.Comm.SendTask(&agent)
	if err != nil {
		var status int
		switch {
		case errors.Is(err, pberrors.ErrInvalidAgentID):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), status)
		return
	}

	// If task is nil, no tasks were found
	if task == nil {
		w.WriteHeader(http.StatusNoContent) // No tasks for this agent
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

// ResultHandler allows sending the results from the agent to the server
func (h *HTTPTransport) ResultHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := h.HandleResult(&result); err != nil {
		var status int
		switch {
		case errors.Is(err, pberrors.ErrInvalidAgentID):
			status = http.StatusBadRequest
		case errors.Is(err, pberrors.ErrInvalidTaskID):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), status)
		return
	}

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
