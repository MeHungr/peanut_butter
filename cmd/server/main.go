package main

import (
	"io"
	"log"
	"os"

	"github.com/MeHungr/peanut-butter/internal/server"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

func main() {
	// Creates a log file if it doesn't exist and opens it in appen mode
	f, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()

	// Sets the logging output to the file and stdout
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	db, err := storage.NewStorage("./pb.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.DB.Close()

	srv := server.New(db, 8080)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server exited: %v", err)
	}
}
