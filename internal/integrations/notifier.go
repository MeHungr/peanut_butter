package integrations

import "github.com/MeHungr/peanut-butter/internal/api"

// Notifier is an interface that defines behavior on agent callback
type Notifier interface {
	SetPayload(payload Payload)
	OnAgentCallback(a *api.Agent) error
}

// Defines the string constants for notifier types
type NotifierString string

const (
	PWNBOARD NotifierString = "pwnboard"
)
