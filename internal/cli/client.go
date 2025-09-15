package cli

import (
	"net/http"
	"time"
)

// Represents the CLI client
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewCLIClient returns a CLI client with default values
func NewCLIClient(baseURL string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		BaseURL:    baseURL,
	}
}
