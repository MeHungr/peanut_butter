// This contains the main server logic for peanut-butter
package server

import (
	"fmt"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/api"
)

var (
	// agents is a package level map that maps agent ids to agents
	agents = make(map[string]api.Agent)
	// tasks is a package level map that maps agent ids to the agent's tasks
	tasks = make(map[string][]api.Task)
)

// Start starts the server and starts listening on the specified port
func Start() {
	// Defines the /register path and uses RegisterHandler to handle data
	http.HandleFunc("/register", RegisterHandler)
	// Defines the /task path and uses TaskHandler to handle data
	http.HandleFunc("/task", TaskHandler)

	// TEMP
	// http.HandleFunc("/enqueue", EnqueueHandler)

	// Starts the server
	err := http.ListenAndServe(":8080", nil)

	// Prints an error if the server fails to start
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
