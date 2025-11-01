package conversion

import (
	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/storage"
	"github.com/MeHungr/peanut-butter/internal/util"
)

func ApiToStorageAgent(a *api.Agent) *storage.Agent {
	return &storage.Agent{
		AgentID:          a.AgentID,
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

func StorageToAPIAgent(a *storage.Agent) *api.Agent {
	return &api.Agent{
		AgentID:          a.AgentID,
		OS:               a.OS,
		Arch:             a.Arch,
		Status:           util.DeriveAgentStatus(*a.LastSeen, a.CallbackInterval),
		Targeted:         a.Targeted,
		AgentIP:          a.AgentIP,
		ServerIP:         a.ServerIP,
		ServerPort:       a.ServerPort,
		CallbackInterval: a.CallbackInterval,
		Hostname:         a.Hostname,
		LastSeen:         a.LastSeen,
	}
}

func StorageToAPITask(t *storage.Task) *api.Task {
	return &api.Task{
		TaskID:    t.TaskID,
		Agent:     api.Agent{AgentID: t.AgentID, OS: t.OS},
		Type:      t.Type,
		Completed: t.Completed,
		Payload:   t.Payload,
		Timeout:   t.Timeout,
		Timestamp: t.Timestamp,
	}
}

func ApiToStorageTask(t *api.Task) *storage.Task {
	return &storage.Task{
		TaskID:    t.TaskID,
		AgentID:   t.AgentID,
		OS:        t.OS,
		Type:      t.Type,
		Completed: t.Completed,
		Payload:   t.Payload,
		Timeout:   t.Timeout,
		Timestamp: t.Timestamp,
	}
}

func ApiToStorageResult(r *api.Result) *storage.Result {
	return &storage.Result{
		ResultID:   r.ResultID,
		AgentID:    r.AgentID,
		OS:         r.OS,
		TaskID:     r.TaskID,
		Output:     r.Output,
		ReturnCode: r.ReturnCode,
		CreatedAt:  r.CreatedAt,
	}
}

func StorageToAPIResult(r *storage.Result) *api.Result {
	return &api.Result{
		ResultID: r.ResultID,
		Task: api.Task{
			TaskID:  r.TaskID,
			Agent:   api.Agent{AgentID: r.AgentID, OS: r.OS},
			Type:    api.TaskType(r.Type),
			Payload: r.Payload,
		},
		Output:     r.Output,
		ReturnCode: r.ReturnCode,
		CreatedAt:  r.CreatedAt,
	}
}

func StoragetoAPIResults(results []storage.Result) []*api.Result {
	apiResults := make([]*api.Result, 0, len(results))
	for _, r := range results {
		apiResults = append(apiResults, StorageToAPIResult(&r))
	}
	return apiResults
}

func ApiToStorageFilter(filter api.AgentFilter) storage.AgentFilter {
	return storage.AgentFilter{
		All:      filter.All,
		IDs:      filter.IDs,
		OSes:     filter.OSes,
		Statuses: filter.Statuses,
	}
}
