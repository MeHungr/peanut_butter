package main

import (
	"time"

	"github.com/MeHungr/peanut-butter/internal/agent"
)

func main() {
	// Constructs the agent and starts it
	//			  (agentID, serverIP, serverPort, callbackInterval, debugMode)
	a := agent.New(agent.GetLocalIP(), "localhost", 8080, 10*time.Second, true)

	a.Start()
}
