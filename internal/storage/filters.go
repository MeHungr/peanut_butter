// Defines filters for querying the database
package storage

// Defines the filter for querying the database for agents
type AgentFilter struct {
	All      bool
	IDs      []string
	OSes     []string
	Statuses []string
}
