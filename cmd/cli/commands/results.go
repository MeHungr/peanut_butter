package commands

import (
	"fmt"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/cli"
	"github.com/spf13/cobra"
)

// resultsCmd represents the results command
var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Manage all results",
	Long:  `Manage all results: list results`,
}

var resultsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List results for agents specified by id or os, or all results",
	Long: `List results for agents specified by id or os, or all results
Example:
	pbctl results list <agent_ids_separated_by_spaces>
	pbctl results list -a | --all
	pbctl results list -o | --os <os>
`,
	Args: requireArgsUnlessFilter("all", "os"),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the wide flag
		wideFlag, err := cmd.Flags().GetBool("wide")
		if err != nil {
			return fmt.Errorf("retrieving wide flag: %w", err)
		}

		// Retrieve the limit flag/value
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return fmt.Errorf("retrieving limit flag: %w", err)
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

		// Retrieve the all flag value
		targetAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			return fmt.Errorf("retriving all flag: %w", err)
		}

		// Grab filter flags and create a filter
		osFilter, err := cmd.Flags().GetStringSlice("os")
		if err != nil {
			return fmt.Errorf("retrieving os flag: %w", err)
		}
		
		// Validate OS input
		if err := validateOSInputs(osFilter); err != nil {
			return fmt.Errorf("validating os: %w", err)
		}

		filter := api.AgentFilter{
			All:  targetAll,
			IDs:  args,
			OSes: osFilter,
		}

		// If watch is enabled, watch
		if interval > 0 {
			cli.Watch(interval, func() error {
				return Client.Results(filter, limit, wideFlag)
			})
			return nil
		}

		// Else, just print the table
		return Client.Results(filter, limit, wideFlag)
	},
}

func init() {
	rootCmd.AddCommand(resultsCmd)

	resultsCmd.AddCommand(resultsListCmd)

	// ===== Flags =====
	resultsListCmd.Flags().StringSliceP("os", "o", []string{}, "Filter results by OS type (accepted: linux, windows, darwin). Singular or comma separated list")
	resultsListCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")
	resultsListCmd.Flags().StringP("watch", "W", "", "Refresh the table periodically (default 2s if no value). Accepts durations like '5', '5s', '500ms'.")
	resultsListCmd.Flags().IntP("limit", "l", 5, "Number of results to display (default 5, use 0 for all)")
	resultsListCmd.Flags().BoolP("all", "a", false, "Override individual IDs and display all results")
	resultsListCmd.Flags().Lookup("watch").NoOptDefVal = "2s"
	// =================
}
