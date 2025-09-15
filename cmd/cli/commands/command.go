package commands

import (
	"github.com/spf13/cobra"
)

var timeoutSeconds int

// commandCmd represents the command command
var commandCmd = &cobra.Command{
	Use:   "command \"<command to execute>\"",
	Short: "Enqueues the following command to targets",
	Long: `Enqueues the following command to assigned targets.
Assign targets using "pbctl targets set|add <comma separated ip addresses>"
Example:
	pbctl command "echo 'Hello, World\!'"`,
	Args: cobra.MinimumNArgs(1), // Ensures at least one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdStr := args[0]

		err := Client.EnqueueCommand(cmdStr, timeoutSeconds)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(commandCmd)

	commandCmd.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 0, "Specify a timeout (in seconds) for the command. 0 = no timeout.")
}
