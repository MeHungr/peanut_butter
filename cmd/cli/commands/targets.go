package commands

import (
	"fmt"

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
	Use:   "add <agent IDs>",
	Short: "Add agents to the current target list",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.AddTargets(client, args)
	},
}

// targetsGetCmd is the 'list' subcommand of targets
var targetsGetCmd = &cobra.Command{
	Use:   "list",
	Short: "List all current targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		wideFlag, err := cmd.Flags().GetBool("wide")
		if err != nil {
			return fmt.Errorf("retrieving wide flag: %w", err)
		}
		return cli.ListTargets(client, wideFlag)
	},
}

// targetsSetCmd is the 'set' subcommand of targets
var targetsSetCmd = &cobra.Command{
	Use:   "set <agent IDs>",
	Short: "Set the current target list to the agents provided",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.SetTargets(client, args)
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
	Use:   "untarget <agent IDs>",
	Short: "Untargets the specified agents",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.Untarget(client, args)
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

	targetsGetCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")

}
