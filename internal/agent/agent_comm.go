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

// Register allows the agent to register with the server
func (a *Agent) Register() error {
	agent := a.Agent
	uri := fmt.Sprintf("http://%s:%d/register", agent.ServerIP, agent.ServerPort)
	var resp api.Message

	// Stores the agent in a RegisterRequest to send to the server
	req := api.RegisterRequest{
		Agent: &agent,
	}

	// POST request with the register request, writes response to resp
	if err := api.DoPost(a.Client, uri, req, &resp); err != nil {
		return fmt.Errorf("Register: %w", err)
	}

	// Print server response
	if a.Debug {
		log.Println(resp.Message)
	}

	return nil
}

// GetTask retrieves a task from the server to be executed
// This function needs to handle its own response codes, so custom logic is needed
func (a *Agent) GetTask() (*api.Task, error) {
	url := fmt.Sprintf("http://%s:%d/task", a.ServerIP, a.ServerPort)

	// Marshals the agent's id into JSON
	body, err := json.Marshal(map[string]string{
		"agent_id": a.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// POST request with the agent's id as the body
	resp, err := a.Post(url, "application/json", bytes.NewBuffer(body))
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
			if a.Debug {
				log.Printf("Agent ID %s no longer recognized, re-registering...", a.ID)
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
func (a *Agent) SendResult(result *api.Result) error {
	uri := fmt.Sprintf("http://%s:%d/result", a.ServerIP, a.ServerPort)
	var resp api.Message

	// Ensures result is not a nil pointer
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Send POST request with result, prints message if debug is on
	if err := api.DoPost(a.Client, uri, result, &resp); err != nil {
		if a.Debug {
			log.Println(resp.Message)
		}
	}

	return nil
}
