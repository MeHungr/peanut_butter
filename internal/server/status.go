// This handles deriving the status of agents (online, stale, offline)
package server

import (
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
)

type StatusMultiplier int

const (
	OnlineMultiplier StatusMultiplier = 2
	StaleMultiplier  StatusMultiplier = 5
)

func DeriveAgentStatus(lastSeen time.Time, now time.Time, callbackInterval time.Duration) api.AgentStatus {
	delta := now.Sub(lastSeen)

	// If agent was never seen, set to offline
	if lastSeen.IsZero() {
		return api.AgentStatusOffline
	}

	// Agent is online if it has called back within 2x the callback interval
	if delta <= callbackInterval*time.Duration(OnlineMultiplier) {
		return api.AgentStatusOnline
		// Agent is stale if it has called back later than 2x but within 5x the callback interval
	} else if delta <= callbackInterval*time.Duration(StaleMultiplier) {
		return api.AgentStatusStale
		// Anything more than 5x the callback interval is offline
	} else {
		return api.AgentStatusOffline
	}
}
