package cli

import (
	"fmt"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/ui"
)

// AddTargets makes agents targets of tasks
func (client *Client) AddTargets(agentIDs []string) error {
	uri := client.BaseURL + "/add-targets"
	targets := api.TargetsRequest{AgentIDs: agentIDs}
	var resp api.Message

	if err := api.DoPost(client.HTTPClient, uri, targets, &resp); err != nil {
		return fmt.Errorf("AddTargets: %w", err)
	}

	// Print the server's response message
	fmt.Println(resp.Message)
	return nil
}

// getTargets retrieves the current targets from the server
func (client *Client) getTargets() ([]*api.Agent, int, error) {
	uri := client.BaseURL + "/get-targets"
	var targetsResponse *api.GetTargetsResponse

	if err := api.DoGet(client.HTTPClient, uri, &targetsResponse); err != nil {
		return nil, 0, fmt.Errorf("getTargets: %w", err)
	}

	return targetsResponse.Agents, targetsResponse.Count, nil
}

// Targets lists the targets from getTargets in a user friendly format
func (client *Client) Targets(wideFlag bool) error {
	targets, count, err := client.getTargets()
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
func (client *Client) Untarget(agentIDs []string) error {
	uri := client.BaseURL + "/untarget"
	var resp api.Message
	targets := api.TargetsRequest{AgentIDs: agentIDs}

	if err := api.DoPost(client.HTTPClient, uri, targets, &resp); err != nil {
		return fmt.Errorf("Untarget: %w", err)
	}

	// Print the server's response message
	fmt.Println(resp.Message)
	return nil
}

// ClearTargets clears all targets
func (client *Client) ClearTargets() error {
	uri := client.BaseURL + "/clear-targets"
	var resp api.Message

	if err := api.DoRequest(client.HTTPClient, http.MethodDelete, uri, nil, &resp); err != nil {
		return fmt.Errorf("ClearTargets: %w", err)
	}

	// Print the response message
	fmt.Println(resp.Message)
	return nil
}

// SetTargets clears all targets then sets the provided agents as targets
func (client *Client) SetTargets(agentIDs []string) error {
	uri := client.BaseURL + "/set-targets"
	targets := api.TargetsRequest{AgentIDs: agentIDs}
	var resp api.Message

	if err := api.DoRequest(client.HTTPClient, http.MethodPut, uri, targets, &resp); err != nil {
		return fmt.Errorf("SetTargets: %w", err)
	}

	// Print the response message
	fmt.Println(resp.Message)
	return nil
}
