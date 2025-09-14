package main

import (
	"log"

	"github.com/MeHungr/peanut-butter/internal/server"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

func main() {
	db, err := storage.NewStorage("./pb.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.DB.Close()

	// Constructs the server and starts it
	// 				 (db object, port)
	srv := server.New(db, 8080)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server exited: %v", err)
	}
}
