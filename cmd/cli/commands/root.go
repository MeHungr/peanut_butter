package commands

import (
	"os"

	"github.com/MeHungr/peanut-butter/internal/cli"
	"github.com/spf13/cobra"
)

// Client defines the CLI client
var Client *cli.Client

// requireArgsUnlessAllFlag returns a function that requires at least one argument unless the all flag is specified
func requireArgsUnlessAllFlag() func(cmd *cobra.Command, args []string) error {
	// Define the function
	return func(cmd *cobra.Command, args []string) error {
		// Parse the selectAll flag from the command
		selectAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		// If selectAll is true:
		if selectAll {
			return nil // skip positional-ID requirement
		}
		// Return the function returned by cobra.MinimumNArgs and pass in cmd and args
		return cobra.MinimumNArgs(1)(cmd, args)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pbctl",
	Short: "pbctl is the CLI for controlling Peanut Butter C2 agents",
	Long: `pbctl is the command-line interface for managing Peanut Butter C2 agents.

You can list connected agents, enqueue commands for execution, manage targets, 
and monitor task results. Use subcommands such as 'agents', 'enqueue', and 'targets' to interact with the server. 
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
	baseURL := "http://localhost:8080"
	Client = cli.NewCLIClient(baseURL)
}
