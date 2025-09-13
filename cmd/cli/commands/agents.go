package commands

import (
	"github.com/MeHungr/peanut-butter/internal/cli"
	"github.com/spf13/cobra"
)

// agentsCmd represents the agents command
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Lists all agents that have connected to the server",
	Long: `Lists the agents that have previously connected to the server.
This command prints them in the format:
AgentID - LastSeenTime`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.ListAgents(client)
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)
}
