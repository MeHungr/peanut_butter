package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Agent struct {
	AgentID          string        `json:"agent_id"`
	ServerIP         string        `json:"server_ip"`
	ServerPort       int           `json:"server_port"`
	CallbackInterval time.Duration `json:"callback_interval,omitempty"`
}

// Register allows the agent to register with the server
func (agent *Agent) Register() error {
	url := fmt.Sprintf("http://%s:%d/register", agent.ServerIP, agent.ServerPort)

	// Marshals the agent id into JSON
	body, err := json.Marshal(map[string]string{
		"agent_id": agent.AgentID,
	})
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// POST request with the agent id as the body
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("Failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// Read the body of the response into a variable
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response:", err)
	}
	//Check response code and handle errors
	if resp.StatusCode == http.StatusOK {
		fmt.Println(string(respBody))
	} else {
		return fmt.Errorf("Server returned status code %d: %s", resp.StatusCode, string(respBody))
	}

	return nil

}

func Start() {
	agent := Agent{
		AgentID: "1",
		ServerIP: "localhost",
		ServerPort: 8080,
	}
	agent.Register()
}
