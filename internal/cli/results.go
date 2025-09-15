package cli

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/ui"
)

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
			CreatedAt:  r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		rows = append(rows, resultRow)
	}
	return rows
}

// getResults returns a list of results for all agents or a specified agent
func (client *Client) getResults(agentID string, limit int) (*api.GetResultsResponse, error) {
	uri := client.BaseURL + "/get-results"
	var out api.GetResultsResponse
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

	if err := api.DoGet(client.HTTPClient, u.String(), &out); err != nil {
		return nil, fmt.Errorf("getResults: %w", err)
	}

	return &out, nil
}

// Results lists the results of agent tasks
func (client *Client) Results(agentID string, limit int, wideFlag bool) error {
	// Retrieves the list of results
	resp, err := client.getResults(agentID, limit)
	if err != nil {
		return fmt.Errorf("Failed to get results: %w", err)
	}

	rows := resultsToRows(resp.Results)
	ui.RenderResults(rows, wideFlag)
	return nil
}
