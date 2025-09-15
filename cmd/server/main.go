package main

import (
	"log"

	"github.com/MeHungr/peanut-butter/internal/server"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

func main() {
	// Initialize the database
	db, err := storage.NewStorage("./pb.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.DB.Close()

	// ========== Config ==========
	serverPort := 8080
	// ============================

	// Constructs the server and starts it
	srv := server.New(db, serverPort)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server exited: %v", err)
	}
}
