// The main server entrypoint
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

	"github.com/MeHungr/peanut-butter/internal/server"
	"github.com/MeHungr/peanut-butter/internal/storage"
	"github.com/kardianos/service"
)

func main() {
	// ========== Config ==========
	serverPort := 80
	integration := true // 3rd party integrations
	// ============================

	// Initialize the database
	db, err := storage.NewStorage("./pb.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.DB.Close()

	// Constructs the server and starts it
	srv := server.New(db, serverPort, integration)

	// Create service wrapper and run as service
	prg := &program{server: srv}
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
