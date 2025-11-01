package main

import (
	"github.com/MeHungr/peanut-butter/cmd/cli/commands"
	"github.com/MeHungr/peanut-butter/internal/cli"
)

func main() {
	// ========== Config ==========
	baseURL := "http://localhost:80"
	// ============================

	// Constructs the client and executes the commands
	commands.Client = cli.NewCLIClient(baseURL)

	commands.Execute()
}
