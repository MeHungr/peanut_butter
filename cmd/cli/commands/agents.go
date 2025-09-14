package commands

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		wideFlag, err := cmd.Flags().GetBool("wide")
		if err != nil {
			return fmt.Errorf("retrieving wide flag: %w", err)
		}
		cli.ListAgents(client, wideFlag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)

	agentsCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")
}
