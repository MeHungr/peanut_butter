package commands

import (
	"fmt"

	"github.com/MeHungr/peanut-butter/internal/api"
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
	Args:  requireArgsUnlessAllFlag(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the all flag value
		targetAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			return fmt.Errorf("retrieving all flag: %w", err)
		}

		return Client.AddTargets(args, targetAll, "")
	},
}

// targetsListCmd is the 'list' subcommand of targets
var targetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all current targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the wide flag
		wideFlag, err := cmd.Flags().GetBool("wide")
		if err != nil {
			return fmt.Errorf("retrieving wide flag: %w", err)
		}

		// Retrieve the watch flag/value
		watchVal, err := cmd.Flags().GetString("watch")
		if err != nil {
			return fmt.Errorf("retrieving watch flag: %w", err)
		}

		// Parse the interval from the watch flag
		interval, err := cli.ParseWatchInterval(watchVal)
		if err != nil {
			return fmt.Errorf("error parsing interval: %w", err)
		}

		// Grab filter flags and create a filter
		osFilter, err := targetsCmd.PersistentFlags().GetString("os")
		if err != nil {
			return fmt.Errorf("retrieving os flag: %w", err)
		}

		filter := api.AgentFilter{
			OSes: parseOSes(osFilter),
		}

		// If watch is enabled, watch
		if interval > 0 {
			cli.Watch(interval, func() error {
				return Client.Targets(wideFlag, filter)
			})
			return nil
		}

		// Else, just print the table
		return Client.Targets(wideFlag, filter)
	},
}

// targetsSetCmd is the 'set' subcommand of targets
var targetsSetCmd = &cobra.Command{
	Use:   "set <agent IDs>",
	Short: "Set the current target list to the agents provided",
	Args:  requireArgsUnlessAllFlag(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the all flag value
		targetAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			return fmt.Errorf("retrieving all flag: %w", err)
		}

		return Client.SetTargets(args, targetAll, "")
	},
}

// targetsClearCmd is the 'clear' subcommand of targets
var targetsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clears the target list",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.ClearTargets()
	},
}

// targetsUntargetCmd is the 'untarget' subcommand of targets
var targetsUntargetCmd = &cobra.Command{
	Use:   "untarget <agent IDs>",
	Short: "Untargets the specified agents",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.Untarget(args)
	},
}

// Adds all subcommands to targets and adds targets to the root command
func init() {
	rootCmd.AddCommand(targetsCmd)
	targetsCmd.AddCommand(targetsAddCmd)
	targetsCmd.AddCommand(targetsListCmd)
	targetsCmd.AddCommand(targetsSetCmd)
	targetsCmd.AddCommand(targetsClearCmd)
	targetsCmd.AddCommand(targetsUntargetCmd)

	// Persistent flags
	targetsCmd.PersistentFlags().StringP("os", "o", "", "Filter or target by OS type (accepted: linux, windows, mac). Singular or comma separated list")

	// Flags
	targetsListCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")
	targetsListCmd.Flags().StringP("watch", "W", "", "Refresh the table periodically (default 2s if no value). Accepts durations like '5', '5s', '500ms'.")
	targetsListCmd.Flags().Lookup("watch").NoOptDefVal = "2s"
	targetsSetCmd.Flags().BoolP("all", "a", false, "Override individual IDs and set all agents to targeted")
	targetsAddCmd.Flags().BoolP("all", "a", false, "Override individual IDs and set all agents to targeted")

	// Hidden flags
	targetsClearCmd.Flags().MarkHidden("os")
}
