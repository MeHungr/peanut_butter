package server

import (
	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

func apiToStorageAgent(a api.Agent) storage.Agent {
    return storage.Agent{
        ID:               a.ID,
        OS:               a.OS,
        Arch:             a.Arch,
        Targeted:         a.Targeted,
        AgentIP:          a.AgentIP,
        ServerIP:         a.ServerIP,
        ServerPort:       a.ServerPort,
        CallbackInterval: a.CallbackInterval, // both are time.Duration
        Hostname:         a.Hostname,
        Status:           a.Status,
        LastSeen:         a.LastSeen,
    }
}

