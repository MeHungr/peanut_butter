// Handles displaying of information in the cli
package ui

import (
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// AgentRow is the cli representation of an agent
type AgentRow struct {
	ID               string
	OS               string
	Arch             string
	Status 
	Targeted         string
	CallbackInterval int // seconds
	Hostname         string
	LastSeen         string
	// Extras
	AgentIP    string
	ServerIP   string
	ServerPort int
}
