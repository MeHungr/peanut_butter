package server

import (
	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

func apiToStorageAgent(a *api.Agent) *storage.Agent {
	return &storage.Agent{
		ID:               a.ID,
		OS:               a.OS,
		Arch:             a.Arch,
		Targeted:         a.Targeted,
		AgentIP:          a.AgentIP,
		ServerIP:         a.ServerIP,
		ServerPort:       a.ServerPort,
		CallbackInterval: a.CallbackInterval,
		Hostname:         a.Hostname,
		LastSeen:         a.LastSeen,
	}
}

func storageToAPIAgent(a *storage.Agent) *api.Agent {
	return &api.Agent{
		ID:               a.ID,
		OS:               a.OS,
		Arch:             a.Arch,
		Targeted:         a.Targeted,
		AgentIP:          a.AgentIP,
		ServerIP:         a.ServerIP,
		ServerPort:       a.ServerPort,
		CallbackInterval: a.CallbackInterval,
		Hostname:         a.Hostname,
		LastSeen:         a.LastSeen,
	}
}

func storageToAPITask(t *storage.Task) *api.Task {
	return &api.Task{
		TaskID:    t.TaskID,
		AgentID:   t.AgentID,
		Type:      t.Type,
		Completed: t.Completed,
		Payload:   t.Payload,
		Timeout:   t.Timeout,
		Timestamp: t.Timestamp,
	}
}

func apiToStorageTask(t *api.Task) *storage.Task {
	return &storage.Task{
		TaskID:    t.TaskID,
		AgentID:   t.AgentID,
		Type:      t.Type,
		Completed: t.Completed,
		Payload:   t.Payload,
		Timeout:   t.Timeout,
		Timestamp: t.Timestamp,
	}
}

func apiToStorageResult(r *api.Result) *storage.Result {
	return &storage.Result{
		ResultID:   r.ResultID,
		AgentID:    r.AgentID,
		TaskID:     r.TaskID,
		Output:     r.Output,
		ReturnCode: r.ReturnCode,
	}
}
