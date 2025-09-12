package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// listAgents reaches out to the /get-agents endpoint and prints connected agents
func listAgents(client *http.Client) error {
	// Sends a GET request to the /get-agents endpoint and handles errors
	resp, err := client.Get("http://localhost:8080/get-agents")
	if err != nil {
		return fmt.Errorf("Failed to send GET request: %w", err)
	}
	defer resp.Body.Close()

	// Decodes JSON into api.Agent slice and handles errors
	var agents []api.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return fmt.Errorf("Failed to decode JSON: %w", err)
	}

	// Prints out each agent and last seen time
	for _, a := range agents {
		fmt.Printf("%s - Last seen: %s\n", a.ID, a.LastSeen.Format(time.RFC3339))
	}

	return nil
}

// Run is the main loop for the cli
func Run() {
	client := &http.Client{Timeout: 10 * time.Second}

	// Opens a new input stream from stdin
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("pbctl > ")
		// Takes the next line of input and handles errors
		if !scanner.Scan() {
			fmt.Println("\nExiting CLI")
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading input:", err)
			}
			return
		}
		input := strings.TrimSpace(scanner.Text())

		switch input {
		case "agents":
			if err := listAgents(client); err != nil {
				fmt.Println("Error:", err)
			}
		case "quit", "exit":
			fmt.Println("Exiting CLI")
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}
