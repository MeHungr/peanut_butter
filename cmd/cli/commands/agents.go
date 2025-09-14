package commands

import (
	"fmt"

	"github.com/MeHungr/peanut-butter/internal/cli"
	"github.com/spf13/cobra"
)

// agentsCmd represents the agents command
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage all agents",
	Long:  `Manage agents: list agents`,
}
var agentsListCmd = &cobra.Command{
	Use: "list",
	Short: "List all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		wideFlag, err := cmd.Flags().GetBool("wide")
		if err != nil {
			return fmt.Errorf("retrieving wide flag: %w", err)
		}
		return cli.ListAgents(client, wideFlag)
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)

	agentsCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")
	agentsCmd.AddCommand(agentsListCmd)
}
