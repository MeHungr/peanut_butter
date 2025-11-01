package commands

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/MeHungr/peanut-butter/internal/cli"
	"github.com/spf13/cobra"
)

// Client defines the CLI client
var Client *cli.Client

// List of valid OSes for the --os flag
var validOSSet = map[string]struct{}{
	"windows": {},
	"linux":   {},
	"freebsd": {},
	"darwin":  {},
}

// keys is a helper function to extract keys from a map
func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	return out
}

// validateOSInputs ensures that an entered OS is a valid option
func validateOSInputs(osList []string) error {
	for _, os := range osList {
		// Make lowercase for case-insensitive match
		os = strings.ToLower(os)
		if _, ok := validOSSet[os]; !ok {
			return fmt.Errorf("invalid OS: %q (valid options: %v)", os, strings.Join(keys(validOSSet), ", "))
		}
	}
	return nil
}

// requireArgsUnlessAllFlag returns a function that requires at least one argument unless the all flag is specified
func requireArgsUnlessFilter(filterFlags ...string) func(cmd *cobra.Command, args []string) error {
	// Define the function
	return func(cmd *cobra.Command, args []string) error {
		// Check if any filter flag has a value
		if slices.ContainsFunc(filterFlags, cmd.Flags().Changed) {
			return nil // skip positional arg requirement
		}

		// If no filter flags provided, require at least one positional arg
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
