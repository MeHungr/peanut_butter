package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	// "os"
	// "strings"
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
	return agents.AgentIDs, nil

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

	// Converts seconds to a duration
	timeoutDuration := time.Duration(timeoutSeconds) * time.Second

	// Retrieve agents from the server
	agents, err := getAgents(client)
	if err != nil {
		fmt.Println(err)
	}

	// Iterate through agents and only send tasks to targeted agents
	for _, agent := range agents {
		if agent.Targeted {
			// Define the task
			task := api.Task{
				Type:      "command",
				AgentID:   agent.ID,
				Completed: false,
				Payload:   cmd,
			}

			// Timeout greater than 0 will be added to the task
			if timeoutSeconds > 0 {
				task.Timeout = &timeoutDuration
			}

			// Marshal into JSON
			taskJSON, err := json.Marshal(task)
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
				return fmt.Errorf("Failed to decode JSON: %w", err)
			}

			fmt.Println(msg.Message)
		}
	}
	return nil
}

// ParseIDs converts a comma-separated string inot a clean slice of agent IDs
func ParseIDs(input string) []string {
	rawIDs := strings.Split(input, ",") // IDs with whitespace

	// Trim whitespace and return new slice
	var ids []string
	for _, id := range rawIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func AddTargets(client *http.Client, agentIDs []string) error {
	url := "http://localhost:8080/add-targets"
	targets := api.AddTargetsRequest{AgentIDs: agentIDs}

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

// // Run is the main loop for the cli
// func Run() {
// 	client := &http.Client{Timeout: 10 * time.Second}
//
// 	// Opens a new input stream from stdin
// 	scanner := bufio.NewScanner(os.Stdin)
//
// 	for {
// 		fmt.Print("pbctl > ")
// 		// Takes the next line of input and handles errors
// 		if !scanner.Scan() {
// 			fmt.Println("\nExiting CLI")
// 			if err := scanner.Err(); err != nil {
// 				fmt.Println("Error reading input:", err)
// 			}
// 			return
// 		}
// 		input := strings.TrimSpace(scanner.Text())
// 		tokens := strings.Fields(input)
// 		cmd := tokens[0]
//
// 		switch cmd {
// 		case "agents":
// 			ListAgents(client)
// 		case "command":
// 			if len(tokens) != 2 {
// 				fmt.Println("Usage:\npbctl > command \"<command + args>\"")
// 				break
// 			}
// 			EnqueueCommand(client, tokens[1], 0)
// 		case "quit", "exit":
// 			fmt.Println("Exiting CLI")
// 			return
// 		default:
// 			fmt.Println("Unknown command")
// 		}
// 	}
// }
