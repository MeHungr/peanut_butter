package main

import (
	"time"

	"github.com/MeHungr/peanut-butter/internal/agent"
)

func main() {
	// ========== Config ==========
	agentID := agent.GetLocalIP()
	serverIP := "localhost"
	serverPort := 8080
	callbackInterval := 10 * time.Second
	debugMode := false
	// ============================

	// Constructs the agent and starts it
	a := agent.New(agentID, serverIP, serverPort, callbackInterval, debugMode)

	a.Start()
}
