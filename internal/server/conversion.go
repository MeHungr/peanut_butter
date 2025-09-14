// Contains conversions between api and database structs
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
		Status:           DeriveAgentStatus(*a.LastSeen, a.CallbackInterval),
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

func storageToAPIResult(r *storage.Result) *api.Result {
	return &api.Result{
		ResultID: r.ResultID,
		Task: api.Task{
			TaskID:  r.TaskID,
			AgentID: r.AgentID,
			Type:    api.TaskType(r.Type),
			Payload: r.Payload,
		},
		Output:     r.Output,
		ReturnCode: r.ReturnCode,
	}
}

func storagetoAPIResults(results []storage.Result) []*api.Result {
	apiResults := make([]*api.Result, 0, len(results))
	for _, r := range results {
		apiResults = append(apiResults, storageToAPIResult(&r))
	}
	return apiResults
}
