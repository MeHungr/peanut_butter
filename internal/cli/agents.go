package cli

import (
	"fmt"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/ui"
)

// agentsToRows converts api.Agents into ui.AgentRows
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

// getAgents returns a list of agents registered with the server
func (client *Client) getAgents() ([]*api.Agent, int, error) {
	uri := client.BaseURL + "/get-agents"
	var agents api.GetAgentsResponse

	if err := api.DoGet(client.HTTPClient, uri, &agents); err != nil {
		return nil, 0, fmt.Errorf("getAgents: %w", err)
	}

	return agents.Agents, agents.Count, nil
}

// Agents reaches out to the /get-agents endpoint and prints connected agents
func (client *Client) Agents(wideFlag bool) error {
	// Retrieves the list of agents
	agents, count, err := client.getAgents()
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
