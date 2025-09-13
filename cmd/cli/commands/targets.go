package commands

import (
	"github.com/MeHungr/peanut-butter/internal/cli"
	"github.com/spf13/cobra"
)

// targetsCmd represents the targets command
var targetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "Manage assigned targets",
	Long:  `Manage assigned targets: add, set, get, or clear target agents`,
}

var targetsAddCmd = &cobra.Command{
	Use:   "add <comma separated agent IDs>",
	Short: "Add agents to the current target list",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentIDs := cli.ParseIDs(args[0])
		return cli.AddTargets(client, agentIDs)
	},
}

var targetsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the list of current targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.ListTargets(client)
	},
}

func init() {
	rootCmd.AddCommand(targetsCmd)
	targetsCmd.AddCommand(targetsAddCmd)
	targetsCmd.AddCommand(targetsGetCmd)
}
