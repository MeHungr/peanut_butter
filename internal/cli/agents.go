package cli

import (
	"fmt"
	"net/url"
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
			ID:               a.AgentID,
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
func (client *Client) getAgents(filter api.AgentFilter) ([]*api.Agent, int, error) {
	uri := client.BaseURL + "/get-agents"
	var agents api.GetAgentsResponse

	// Create a url to add a query to
	fullURL, err := url.Parse(uri)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid base URL: %w", err)
	}

	// Create the query
	query := fullURL.Query()

	// Add query parameters
	if filter.All {
		query.Set("all", "true")
	}
	for _, id := range filter.IDs {
		query.Add("id", id)
	}
	for _, os := range filter.OSes {
		query.Add("os", os)
	}
	for _, status := range filter.Statuses {
		query.Add("status", status)
	}

	// Encode the query into the url
	fullURL.RawQuery = query.Encode()

	if err := api.DoGet(client.HTTPClient, fullURL.String(), &agents); err != nil {
		return nil, 0, fmt.Errorf("getAgents: %w", err)
	}

	return agents.Agents, agents.Count, nil
}

// Agents reaches out to the /get-agents endpoint and prints connected agents
func (client *Client) Agents(wideFlag bool, filter api.AgentFilter) error {
	// Retrieves the list of agents
	agents, count, err := client.getAgents(filter)
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
