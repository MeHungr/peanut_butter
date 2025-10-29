// agent_comm.go defines behavior for agent communication and determines
// which transport to use when communicating with the server
package agent

import "github.com/MeHungr/peanut-butter/internal/api"

// CommManager manages the transports of an agent and determines which one to use
type CommManager struct {
	Transports map[string]Transport // "http" -> HTTPTransport
	Active     Transport            // Currently used transport
}

// Register allows the agent to register with the server
func (cm *CommManager) Register(a *api.Agent, debug bool) error {
	return cm.Active.Register(a, debug)
}

// GetTask retrieves a task from the server to be executed
func (cm *CommManager) GetTask(a *api.Agent, debug bool) (*api.Task, error) {
	return cm.Active.GetTask(a, debug)
}

// SendResult sends a result from an agent to the server
func (cm *CommManager) SendResult(a *api.Agent, result *api.Result, debug bool) error {
	return cm.Active.SendResult(a, result, debug)
}
