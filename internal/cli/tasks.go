package cli

import (
	"fmt"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// EnqueueCommand sends a task for each targeted agent to the server with type command and the specified payload
func (client *Client) EnqueueCommand(cmd string, timeoutSeconds int) error {
	uri := client.BaseURL + "/enqueue"
	var resp api.Message
	req := api.EnqueueRequest{
		Type:    api.Command,
		Payload: cmd,
		Timeout: timeoutSeconds,
	}

	if err := api.DoPost(client.HTTPClient, uri, req, &resp); err != nil {
		return fmt.Errorf("EnqueueCommand: %w", err)
	}

	fmt.Println(resp.Message)

	return nil
}
