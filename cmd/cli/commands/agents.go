package commands

import (
	"fmt"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/cli"
	"github.com/spf13/cobra"
)

// agentsCmd represents the agents command
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage all agents",
	Long:  `Manage agents: list agents`,
}

// agentsListCmd represents the agents 'list' subcommand
var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agents",
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

		osFilter, err := cmd.Flags().GetStringSlice("os")
		if err != nil {
			return fmt.Errorf("retrieving os flag: %w", err)
		}

		filter := api.AgentFilter{
			OSes: osFilter,
		}

		// Parse the interval from the watch flag
		interval, err := cli.ParseWatchInterval(watchVal)
		if err != nil {
			return fmt.Errorf("error parsing interval: %w", err)
		}

		// If watch is enabled, watch
		if interval > 0 {
			cli.Watch(interval, func() error {
				return Client.Agents(wideFlag, filter)
			})
			return nil
		}

		// Else, just print the table
		return Client.Agents(wideFlag, filter)
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)

	agentsCmd.AddCommand(agentsListCmd)

	// Flags
	agentsListCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")
	agentsListCmd.Flags().StringP("watch", "W", "", "Refresh the table periodically (default 2s if no value). Accepts durations like '5', '5s', '500ms'.")
	agentsListCmd.Flags().Lookup("watch").NoOptDefVal = "2s"
	agentsListCmd.Flags().StringSliceP("os", "o", []string{}, "Filter or target by OS type (accepted: linux, windows, freebsd, mac). Singular or comma separated list")

}
