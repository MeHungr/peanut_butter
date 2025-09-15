package commands

import (
	"fmt"

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
	Short: "List all results or all results for a given agent",
	Long: `List all results or all results for a given agent
Example:
	pbctl results list
	pbctl results list -a agent1
	pbctl results list --agent agent1
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the agent id flag
		agentID, err := cmd.Flags().GetString("agent")
		if err != nil {
			return fmt.Errorf("retrieving agent flag: %w", err)
		}

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

		// If watch is enabled, watch
		if interval > 0 {
			cli.Watch(interval, func() error {
				return Client.Results(agentID, limit, wideFlag)
			})
			return nil
		}

		// Else, just print the table
		return Client.Results(agentID, limit, wideFlag)
	},
}

func init() {
	rootCmd.AddCommand(resultsCmd)

	resultsCmd.AddCommand(resultsListCmd)

	resultsListCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")
	resultsListCmd.Flags().StringP("watch", "W", "", "Refresh the table periodically (default 2s if no value). Accepts durations like '5', '5s', '500ms'.")
	resultsListCmd.Flags().IntP("limit", "l", 5, "Number of results to display (default 5, use 0 for all)")
	resultsListCmd.Flags().StringP("agent", "a", "", "Filter results by agent ID")
	resultsListCmd.Flags().Lookup("watch").NoOptDefVal = "2s"
	resultsListCmd.Flags().Lookup("agent").NoOptDefVal = ""
}
