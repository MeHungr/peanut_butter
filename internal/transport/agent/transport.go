package agent

import "github.com/MeHungr/peanut-butter/internal/api"

// Transport defines methods for communicating with a server
type Transport interface {
	Register(*api.Agent, bool) error
	GetTask(*api.Agent, bool) (*api.Task, error)
	SendResult(*api.Agent, *api.Result, bool) error
}
