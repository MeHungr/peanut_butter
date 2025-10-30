// The main agent entrypoint
// This can be run with no arguments to run in the foreground,
// or, it accepts arguments for running as a service.
// Valid arguments:
//
//	install
//	uninstall
//	start
//	stop
package main

import (
	"log"
	"os"
	"time"

	"github.com/MeHungr/peanut-butter/internal/agent"
	"github.com/kardianos/service"
)

func main() {
	// ========== Config ==========
	agentID := agent.GetLocalIP()
	serverIP := "10.64.36.58"
	serverPort := 8080
	callbackInterval := 5 * time.Minute
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

	// Allows control functions
	if len(os.Args) > 1 {
		err = service.Control(svc, os.Args[1])
		if err != nil {
			log.Fatalf("Valid actions: install, uninstall, start, stop. Error: %v", err)
		}
		return
	}

	// svc.Run will block and run in foreground if no arguments are provided
	if err := svc.Run(); err != nil {
		log.Fatalf("service run error: %v", err)
	}
}
