package agent

import (
	"log"
	"net"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

// GetLocalIP returns the local ip of the agent
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "?.?.?.?"
	}
	for _, addr := range addrs {
		// Filters out loopback addresses
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "?.?.?.?"
}

// Converts the agent.Agent struct into an api.Agent
func (a *Agent) ToAPI() api.Agent {
	return api.Agent{
		ID:               a.ID,
		OS:               a.OS,
		Arch:             a.Arch,
		AgentIP:          a.AgentIP,
		ServerIP:         a.ServerIP,
		ServerPort:       a.ServerPort,
		CallbackInterval: a.CallbackInterval,
		Hostname:         a.Hostname,
		LastSeen:         a.LastSeen,
	}
}

// registerUntilDone has the agent attempt to register with the server until it is accepted
func (a *Agent) registerUntilDone() {
	for {
		if err := a.Register(); err != nil {
			if a.Debug {
				log.Println(err)
			}
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}
