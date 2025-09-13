package commands

import (
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var client = &http.Client{Timeout: 10 * time.Second}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pbctl",
	Short: "pbctl is the CLI for controlling Peanut Butter C2 agents",
	Long: `pbctl is the command-line interface for managing Peanut Butter C2 agents.

You can list connected agents, enqueue commands for execution, manage targets, 
and monitor task results. Use subcommands such as 'agents', 'enqueue', 'set-targets',
and 'console' to interact with the server. 
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
