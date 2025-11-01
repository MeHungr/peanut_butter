package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/pberrors"
)

// HTTPSTransport defines the struct for HTTPS communication
type HTTPSTransport struct {
	Client *http.Client
}

// Register allows the agent to register with the server
func (h *HTTPSTransport) Register(agent *api.Agent, debug bool) error {
	uri := fmt.Sprintf("https://%s:%d/register", agent.ServerIP, agent.ServerPort)
	var resp api.Message

	// Stores the agent in a RegisterRequest to send to the server
	req := api.RegisterRequest{
		Agent: agent,
	}

	// POST request with the register request, writes response to resp
	if err := api.DoPost(h.Client, uri, req, &resp); err != nil {
		return fmt.Errorf("Register: %w", err)
	}

	// Print server response
	if debug {
		log.Println(resp.Message)
	}

	return nil
}

// GetTask retrieves a task from the server to be executed
// This function needs to handle its own response codes, so custom logic is needed
func (h *HTTPSTransport) GetTask(a *api.Agent, debug bool) (*api.Task, error) {
	url := fmt.Sprintf("https://%s:%d/task", a.ServerIP, a.ServerPort)

	// Marshals the agent's id into JSON
	body, err := json.Marshal(map[string]string{
		"agent_id": a.AgentID,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// POST request with the agent's id as the body
	resp, err := h.Client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("Failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// Read the body of the response
	respBody, _ := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil, nil // No task available

	case http.StatusBadRequest:
		// If agent ID is unregistered, reregister
		if strings.Contains(string(respBody), "Invalid agent ID") {
			if debug {
				log.Printf("Agent ID %s no longer recognized, re-registering...", a.AgentID)
			}
			return nil, pberrors.ErrInvalidAgentID
		} else { // Else, throw error for bad request
			return nil, fmt.Errorf("Server returned status %d: %s", resp.StatusCode, string(respBody))
		}

	case http.StatusOK:
		// Decodes the JSON body into a Task
		var agentTask api.Task
		if err := json.Unmarshal(respBody, &agentTask); err != nil {
			return nil, fmt.Errorf("Failed to decode task JSON: %w", err)
		}
		return &agentTask, nil

	default:
		// Throw other errors
		return nil, fmt.Errorf("Server returned status %d: %s", resp.StatusCode, string(respBody))
	}
}

// SendResult sends a result from an agent to the server
func (h *HTTPSTransport) SendResult(a *api.Agent, result *api.Result, debug bool) error {
	uri := fmt.Sprintf("https://%s:%d/result", a.ServerIP, a.ServerPort)
	var resp api.Message

	// Ensures result is not a nil pointer
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Send POST request with result, prints message if debug is on
	if err := api.DoPost(h.Client, uri, result, &resp); err != nil {
		if debug {
			log.Println(resp.Message)
		}
	}

	return nil
}
