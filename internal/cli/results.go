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
			AgentOS:    r.OS,
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
func (client *Client) getResults(filter api.AgentFilter, limit int) (*api.GetResultsResponse, error) {
	uri := client.BaseURL + "/get-results"
	var out api.GetResultsResponse

	// Create a url to add a query to
	fullURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Create the query
	query := fullURL.Query()

	// Add query parameters
	if filter.All {
		query.Set("all", "true")
	}
	for _, id := range filter.IDs {
		query.Add("agent_id", id)
	}
	for _, os := range filter.OSes {
		query.Add("os", os)
	}
	for _, status := range filter.Statuses {
		query.Add("status", status)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	// Encode the query into the url
	fullURL.RawQuery = query.Encode()

	if err := api.DoGet(client.HTTPClient, fullURL.String(), &out); err != nil {
		return nil, fmt.Errorf("getResults: %w", err)
	}

	return &out, nil
}

// Results lists the results of agent tasks
func (client *Client) Results(filter api.AgentFilter, limit int, wideFlag bool) error {
	// Retrieves the list of results
	resp, err := client.getResults(filter, limit)
	if err != nil {
		return fmt.Errorf("Failed to get results: %w", err)
	}

	rows := resultsToRows(resp.Results)
	ui.RenderResults(rows, wideFlag)
	return nil
}
