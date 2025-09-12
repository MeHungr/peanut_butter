package main

import (
	"net/http"
	"time"

	"github.com/MeHungr/peanut-butter/internal/agent"
	"github.com/MeHungr/peanut-butter/internal/api"
)

func main() {
	a := agent.Agent{
		Agent: api.Agent{
			ID:               agent.GetLocalIP(),
			ServerIP:         "localhost",
			ServerPort:       8080,
			CallbackInterval: 10 * time.Second,
		},
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	a.Start()
}
