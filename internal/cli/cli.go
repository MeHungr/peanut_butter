// Contains logic for executing commands via the cli
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/ui"
)

var baseURL = "http://localhost:8080"

// humanizeSince humanizes time deltas into a user friendly format
func humanizeSince(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	delta := time.Since(t)

	switch {
	// In the case of clock skew, allow for future case
	case delta < 0:
		return "in the future"
	// Now
	case delta < time.Second:
		return "now"
	// Seconds
	case delta < time.Minute:
		return fmt.Sprintf("%ds ago", int(delta.Seconds()))
	// Minutes
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	// Hours
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	// Anything else
	default:
		return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
	}
}

// boolToString converts a boolean value to 'yes' or 'no'
func boolToString(b bool) string {
	var str string
	switch b {
	case true:
		str = "yes"
	case false:
		str = "no"
	}
	return str

}

func agentsToRows(agents []*api.Agent) []ui.AgentRow {
	rows := make([]ui.AgentRow, 0, len(agents))

	// Iterate through agents and convert to AgentRows
	for _, a := range agents {
		agentRow := ui.AgentRow{
			ID:               a.ID,
			OS:               a.OS,
			Arch:             a.Arch,
			Status:           string(a.Status),
			Targeted:         boolToString(a.Targeted),
			Hostname:         a.Hostname,
			LastSeen:         humanizeSince(*a.LastSeen),
			CallbackInterval: int(a.CallbackInterval / time.Second),
			AgentIP:          a.AgentIP,
			ServerIP:         a.ServerIP,
			ServerPort:       a.ServerPort,
		}
		rows = append(rows, agentRow)
	}
	return rows
}

func resultsToRows(results []*api.Result) []ui.ResultRow {
	rows := make([]ui.ResultRow, 0, len(results))

	// Iterate through agents and convert to AgentRows
	for _, r := range results {
		resultRow := ui.ResultRow{
			ResultID:   strconv.Itoa(r.ResultID),
			TaskID:     strconv.Itoa(r.TaskID),
			AgentID:    r.AgentID,
			Type:       string(r.Type),
			Output:     r.Output,
			Payload:    r.Payload,
			ReturnCode: strconv.Itoa(r.ReturnCode),
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		rows = append(rows, resultRow)
	}
	return rows
}

// getAgents returns a list of agents registered with the server
func getAgents(client *http.Client) ([]*api.Agent, int, error) {
	url := baseURL + "/get-agents"
	// Sends a GET request to the /get-agents endpoint and handles errors
	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to send GET request: %w", err)
	}
	defer resp.Body.Close()

	// Decodes JSON into api.Agent slice and handles errors
	var agents api.GetAgentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, 0, fmt.Errorf("Failed to decode JSON: %w", err)
	}

	// Returns the list of agents
	return agents.Agents, agents.Count, nil

}

// getResults returns a list of results for all agents or a specified agent
func getResults(client *http.Client, agentID string, limit int) (*api.GetResultsResponse, error) {
	uri := baseURL + "/get-results"
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	if agentID != "" {
		q.Set("agent_id", agentID)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("Failed to send GET: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server returned status %d: %s", resp.StatusCode, respBody)
	}

	var out api.GetResultsResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal JSON: %w", err)
	}

	return &out, nil
}

// clearScreen clears the screen on multiple OSes
func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

// ParseWatchInterval parses a watch flag string into a duration.
// If val is empty, it returns 0 (disabled).
// If val is a bare number, it's interpreted as seconds.
func ParseWatchInterval(val string) (time.Duration, error) {
	if val == "" {
		return 0, nil
	}
	// try full duration format first: 500ms, 2s, 1m
	if d, err := time.ParseDuration(val); err == nil {
		return d, nil
	}
	// try interpreting as seconds if just a number
	if d, err := time.ParseDuration(val + "s"); err == nil {
		return d, nil
	}
	return 0, fmt.Errorf("invalid watch interval: %q (examples: 2s, 5, 750ms)", val)
}

// Watch repeatedly clears the screen and executes fn at a given interval.
func Watch(interval time.Duration, fn func() error) {
	for {
		clearScreen()
		if err := fn(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		time.Sleep(interval)
	}
}

// ListAgents reaches out to the /get-agents endpoint and prints connected agents
func ListAgents(client *http.Client, wideFlag bool) error {
	// Retrieves the list of agents
	agents, count, err := getAgents(client)
	if err != nil {
		return err
	}

	switch count {
	case 1:
		fmt.Printf("%d agents found\n", count)
	default:
		fmt.Printf("%d agents found \n", count)
	}

	rows := agentsToRows(agents)
	ui.RenderAgents(rows, wideFlag)

	return nil
}

func ListResults(client *http.Client, agentID string, limit int, wideFlag bool) error {
	// Retrieves the list of results
	resp, err := getResults(client, agentID, limit)
	if err != nil {
		return fmt.Errorf("Failed to get results: %w", err)
	}

	rows := resultsToRows(resp.Results)
	ui.RenderResults(rows, wideFlag)
	return nil
}

// EnqueueCommand sends a task for each targeted agent to the server with type command and the specified payload
func EnqueueCommand(client *http.Client, cmd string, timeoutSeconds int) error {
	url := baseURL + "/enqueue"

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
	url := baseURL + "/add-targets"
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
	url := baseURL + "/get-targets"

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
func ListTargets(client *http.Client, wideFlag bool) error {
	targets, count, err := getTargets(client)
	if err != nil {
		return err
	}

	switch count {
	case 1:
		fmt.Printf("%d target found\n", count)
	default:
		fmt.Printf("%d targets found \n", count)
	}
	rows := agentsToRows(targets)
	ui.RenderAgents(rows, wideFlag)

	return nil
}

// Untarget sets a list of agents as untargeted
func Untarget(client *http.Client, agentIDs []string) error {
	url := baseURL + "/untarget"

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
	url := baseURL + "/clear-targets"

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
	url := baseURL + "/set-targets"

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
