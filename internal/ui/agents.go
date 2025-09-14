// Handles displaying of information in the cli
package ui

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	// "github.com/jedib0t/go-pretty/v6/text"
)

// AgentRow is the cli representation of an agent
type AgentRow struct {
	ID       string
	OS       string
	Arch     string
	Status   string
	Targeted string
	Hostname string
	LastSeen string
	// Extras
	CallbackInterval int // seconds
	AgentIP          string
	ServerIP         string
	ServerPort       int
}

// RenderAgents renders a table of agents to the command line
func RenderAgents(rows []AgentRow, wide bool) {
	// Create a new table directed to stdout
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)

	// Create the header and append it to the table
	header := table.Row{"ID", "TARGETED", "STATUS", "LAST SEEN", "HOSTNAME", "OS/ARCH"}
	if wide {
		header = append(header, "CALLBACK INTERVAL", "AGENT IP", "SERVER IP", "SERVER PORT")
	}
	tw.AppendHeader(header)

	// Create the rows and append them to the table
	for _, r := range rows {
		row := table.Row{r.ID, r.Targeted, r.Status, r.LastSeen, r.Hostname, r.OS + "/" + r.Arch}
		if wide {
			row = append(row, r.CallbackInterval, r.AgentIP, r.ServerIP, r.ServerPort)
		}
		tw.AppendRow(row)
	}

	// Render the table
	tw.Render()
}
