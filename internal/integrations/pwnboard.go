package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// PwnboardNotifier implements the Notifier interface and
// is used to notify pwnboard of a callback
type PwnboardNotifier struct {
	PwnboardURL string
	payload     Payload
	Client      *http.Client
}

// Payload represents the json payload to send to pwnboard
type Payload struct {
	AgentIP     string `json:"ip"`
	Application string `json:"application"`
	AccessType  string `json:"access_type"`
}

// OnAgentCallback sends a json payload to pwnboard
func (pn *PwnboardNotifier) OnAgentCallback(a *api.Agent) error {
	jsonPayload, err := json.Marshal(pn.payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}

	resp, err := pn.Client.Post(pn.PwnboardURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("Error during POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp != nil {
		var respStr string
		json.NewDecoder(resp.Body).Decode(&respStr)
		log.Printf("Response from pwnboard: %s\n", respStr)
	}

	return nil
}

func (pn *PwnboardNotifier) SetPayload(payload Payload) {
	pn.payload = payload
}
