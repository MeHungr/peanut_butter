// This defines the setup of the server as a system service.
// The logic is separated to allow for running the executable
// standalone as well as as a service
package main

import (
	"log"

	"github.com/MeHungr/peanut-butter/internal/server"
	"github.com/kardianos/service"
)

func getServiceConfig() *service.Config {
	svcName := "pb-server"
	svcDisplayName := "Peanut Butter C2 Server"
	svcDesc := "The server for the Peanut Butter C2 framework"
	return &service.Config{
		Name:        svcName,
		DisplayName: svcDisplayName,
		Description: svcDesc,
	}
}

type program struct {
	server *server.Server
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Run agent in background
	go p.run()
	return nil
}

func (p *program) run() {
	// Calls the agent start method
	p.server.Start()
}

func (p *program) Stop(s service.Service) error {
	// Process will exit when killed. If debug is enabled, print msg
	log.Println("Stopping server service...")
	return nil
}
