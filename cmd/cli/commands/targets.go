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

// targetsAddCmd is the 'add' subcommand of targets
var targetsAddCmd = &cobra.Command{
	Use:   "add <comma separated agent IDs>",
	Short: "Add agents to the current target list",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentIDs := cli.ParseIDs(args[0])
		return cli.AddTargets(client, agentIDs)
	},
}

// targetsGetCmd is the 'get' subcommand of targets
var targetsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the list of current targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.ListTargets(client)
	},
}

// targetsSetCmd is the 'set' subcommand of targets
var targetsSetCmd = &cobra.Command{
	Use:   "set <comma separated agent IDs>",
	Short: "Set the current target list to the agents provided",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentIDs := cli.ParseIDs(args[0])
		return cli.SetTargets(client, agentIDs)
	},
}

// targetsClearCmd is the 'clear' subcommand of targets
var targetsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clears the target list",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.ClearTargets(client)
	},
}

// targetsUntargetCmd is the 'untarget' subcommand of targets
var targetsUntargetCmd = &cobra.Command{
	Use:   "untarget <comma separated agent IDs>",
	Short: "Untargets the specified agents",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentIDs := cli.ParseIDs(args[0])
		return cli.Untarget(client, agentIDs)
	},
}

// Adds all subcommands to targets and adds targets to the root command
func init() {
	rootCmd.AddCommand(targetsCmd)
	targetsCmd.AddCommand(targetsAddCmd)
	targetsCmd.AddCommand(targetsGetCmd)
	targetsCmd.AddCommand(targetsSetCmd)
	targetsCmd.AddCommand(targetsClearCmd)
	targetsCmd.AddCommand(targetsUntargetCmd)
}
