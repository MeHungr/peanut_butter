package server

import "github.com/MeHungr/peanut-butter/internal/api"

// Transport defines methods for communicating with an agent
type Transport interface {
	Start() error
	RegisterAgent(*api.Agent) error
	SendTask(*api.Task) (*api.Task, error)
	HandleResult(*api.Result) error
}
