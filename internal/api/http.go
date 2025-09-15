package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DoGet sends a GET request and unmarshals JSON into out
func DoGet(client *http.Client, uri string, outPtr any) error {
	// Make the get request
	resp, err := client.Get(uri)
	if err != nil {
		return fmt.Errorf("GET %s failed: %w", uri, err)
	}
	defer resp.Body.Close()

	// If not ok, error and print response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s failed: status %d: %s", uri, resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(outPtr)
}

// DoPost sends a POST request and unmarshals JSON into out
func DoPost(client *http.Client, uri string, body any, outPtr any) error {
	return DoRequest(client, http.MethodPost, uri, body, outPtr)
}

// DoRequest is a generic helper for POST/PUT/DELETE with a JSON body
func DoRequest(client *http.Client, method, uri string, body any, outPtr any) error {
	// Create a buffer to pass into the request
	var buf io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body failed %w", err)
		}
		buf = bytes.NewBuffer(data)
	}

	// Form the request
	req, err := http.NewRequest(method, uri, buf)
	if err != nil {
		return fmt.Errorf("build request failed %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s failed: %w", method, uri, err)
	}
	defer resp.Body.Close()

	// Check for bad status codes and error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s failed: status %d: %s", method, uri, resp.StatusCode, b)
	}

	// Write the body to out
	if outPtr != nil {
		return json.NewDecoder(resp.Body).Decode(outPtr)
	}
	return nil
}
