// Contains logic for executing commands via the cli
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// getAgents returns a list of agents registered with the server
func getAgents(client *http.Client) ([]*api.Agent, error) {
	url := "http://localhost:8080/get-agents"
	// Sends a GET request to the /get-agents endpoint and handles errors
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to send GET request: %w", err)
	}
	defer resp.Body.Close()

	// Decodes JSON into api.Agent slice and handles errors
	var agents api.GetAgentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, fmt.Errorf("Failed to decode JSON: %w", err)
	}

	// Returns the list of agents
	return agents.Agents, nil

}

// ListAgents reaches out to the /get-agents endpoint and prints connected agents
func ListAgents(client *http.Client) error {
	// Retrieves the list of agents
	agents, err := getAgents(client)
	if err != nil {
		return err
	}

	// Prints out each agent and last seen time
	for _, a := range agents {
		fmt.Printf("%s - Last seen: %s\n", a.ID, a.LastSeen.Format(time.RFC3339))
	}

	return nil
}

// EnqueueCommand sends a task for each targeted agent to the server with type command and the specified payload
func EnqueueCommand(client *http.Client, cmd string, timeoutSeconds int) error {
	url := "http://localhost:8080/enqueue"

	req := api.EnqueueRequest{
		Type:    api.Command,
		Payload: cmd,
		Timeout: timeoutSeconds,
	}

	// Marshal into JSON
	taskJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// POST the JSON payload
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(taskJSON))
	if err != nil {
		return fmt.Errorf("Failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	// Checks the status code and reports errors
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// If the status is 201 Created, prints the message
	var msg api.Message
	if err := json.Unmarshal(respBody, &msg); err != nil {
		return fmt.Errorf("Failed to unmarshal JSON: %w", err)
	}

	fmt.Println(msg.Message)

	return nil
}

// AddTargets makes agents targets of tasks
func AddTargets(client *http.Client, agentIDs []string) error {
	url := "http://localhost:8080/add-targets"
	targets := api.TargetsRequest{AgentIDs: agentIDs}

	payload, err := json.Marshal(targets)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("Failed to send POST: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// getTargets retrieves the current targets from the server
func getTargets(client *http.Client) ([]*api.Agent, int, error) {
	url := "http://localhost:8080/get-targets"

	var targetsResponse *api.GetTargetsResponse

	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to send GET: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&targetsResponse); err != nil {
		return nil, 0, fmt.Errorf("Failed to decode JSON: %w", err)
	}

	return targetsResponse.Agents, targetsResponse.Count, nil
}

// ListTargets lists the targets from getTargets in a user friendly format
func ListTargets(client *http.Client) error {
	targets, count, err := getTargets(client)
	if err != nil {
		return err
	}

	fmt.Printf("Number of targets: %d\n\n", count)
	for _, target := range targets {
		fmt.Printf("%s - Last Seen: %s\n", target.ID, target.LastSeen.Format(time.RFC3339))
	}

	return nil
}

// Untarget sets a list of agents as untargeted
func Untarget(client *http.Client, agentIDs []string) error {
	url := "http://localhost:8080/untarget"

	// Marshal the agent ids into json
	payload, err := json.Marshal(api.TargetsRequest{AgentIDs: agentIDs})
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// Send the POST request
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("Failed to send POST: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, respBody)
	}

	var msg api.Message
	if err := json.Unmarshal(respBody, &msg); err != nil {
		return fmt.Errorf("Failed to unmarshal JSON: %w", err)
	}

	fmt.Println(msg.Message)

	return nil
}

// ClearTargets clears all targets
func ClearTargets(client *http.Client) error {
	url := "http://localhost:8080/clear-targets"

	// Construct DELETE request
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("Failed to build request: %w", err)
	}

	// Receive response
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send DELETE: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned status %d: %s", resp.StatusCode, respBody)
	}

	// Unmarshal the JSON
	var msg api.Message
	if err := json.Unmarshal(respBody, &msg); err != nil {
		return fmt.Errorf("Failed to unmarshal JSON: %w", err)
	}

	// Print the response message
	fmt.Println(msg.Message)

	return nil
}

// SetTargets clears all targets then sets the provided agents as targets
func SetTargets(client *http.Client, agentIDs []string) error {
	url := "http://localhost:8080/set-targets"

	// Construct PUT payload
	payload, err := json.Marshal(api.TargetsRequest{AgentIDs: agentIDs})
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	// Construct PUT request
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("Failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send PUT: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned status %d: %s", resp.StatusCode, respBody)
	}

	// Unmarshal the JSON
	var msg api.Message
	if err := json.Unmarshal(respBody, &msg); err != nil {
		return fmt.Errorf("Failed to unmarshal JSON: %w", err)
	}

	// Print the response message
	fmt.Println(msg.Message)

	return nil
}
