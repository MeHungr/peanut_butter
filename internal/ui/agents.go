// Handles displaying of information in the cli
package ui

import (
	"fmt"
	"os"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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

// ResultRow is the cli representation of a result
type ResultRow struct {
	ResultID   string
	TaskID     string
	AgentID    string
	Output     string
	ReturnCode string
	Payload    string
	Type       string
	CreatedAt  string
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

	// Style options
	tw.SetColumnConfigs([]table.ColumnConfig{
		{
			Name: "STATUS",
			Transformer: func(val any) string {
				s := fmt.Sprint(val)
				switch s {
				case string(api.AgentStatusOnline):
					return text.Colors{text.FgGreen}.Sprint(s)
				case string(api.AgentStatusStale):
					return text.Colors{text.FgYellow}.Sprint(s)
				case string(api.AgentStatusOffline):
					return text.Colors{text.FgRed}.Sprint(s)
				default:
					return s
				}
			},
		},
	})
	tw.SetStyle(table.StyleColoredMagentaWhiteOnBlack)
	// Render the table
	tw.Render()
}

func RenderResults(rows []ResultRow, wide bool) {
	// Create a new table directed to stdout
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)

	// Create the header and append it to the table
	header := table.Row{"AGENT ID", "TYPE", "PAYLOAD", "OUTPUT"}
	if wide {
		header = append(header, "RETURN CODE", "RESULT ID", "TASK ID", "CREATED AT")
	}
	tw.AppendHeader(header)

	// Create the rows and append them to the table
	for _, r := range rows {
		row := table.Row{r.AgentID, r.Type, r.Payload, r.Output}
		if wide {
			row = append(row, r.ReturnCode, r.ResultID, r.TaskID, r.CreatedAt)
		}
		tw.AppendRow(row)
	}

	// Allows the OUTPUT column to wrap
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "PAYLOAD", WidthMax: 40},
		{Name: "OUTPUT", WidthMax: 60},
	})

	tw.SetStyle(table.StyleColoredMagentaWhiteOnBlack)
	tw.Render()
}
