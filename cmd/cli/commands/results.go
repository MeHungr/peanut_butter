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
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var resultsListCmd = &cobra.Command{
	Use:   "list <all|agent_id>",
	Short: "List all results or all results for a given agent",
	Long: `List all results or all results for a given agent
Example:
	pbctl results list all
	pbctl results list agent1
`,
	Args: cobra.MinimumNArgs(1), // Ensures at least one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse the agent id. Set to "" if "all" is passed
		agentID := args[0]
		if agentID == "all" {
			agentID = ""
		}
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

		// If watch is enabled, watch
		if interval > 0 {
			cli.Watch(interval, func() error {
				return cli.ListResults(client, agentID, wideFlag)
			})
			return nil
		}

		// Else, just print the table
		return cli.ListResults(client, agentID, wideFlag)
	},
}

func init() {
	rootCmd.AddCommand(resultsCmd)

	resultsCmd.AddCommand(resultsListCmd)

	resultsListCmd.Flags().BoolP("wide", "w", false, "Show more columns in the table")
	resultsListCmd.Flags().StringP("watch", "W", "", "Refresh the table periodically (default 2s if no value). Accepts durations like '5', '5s', '500ms'.")
	resultsListCmd.Flags().Lookup("watch").NoOptDefVal = "2s"
}
