// This contains the main server logic for peanut-butter
package server

import (
	"fmt"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

var (
	agents = make(map[string]*api.Agent)  // package level map that maps agent ids to agents
	tasks  = make(map[string][]*api.Task) // package level map that maps agent ids to the agent's tasks
)

type Server struct {
	storage *storage.Storage
	port    int
}

func New(storage *storage.Storage, port int) *Server {
	return &Server{
		storage: storage,
		port:    port,
	}
}

// Start starts the server and starts listening on the specified port
func (srv *Server) Start() error {

	// --------------------------------
	// AGENT ENDPOINTS
	// --------------------------------
	// Defines the /register path and uses RegisterHandler to handle data
	http.HandleFunc("/register", srv.RegisterHandler)
	// Defines the /task path and uses TaskHandler to handle data
	http.HandleFunc("/task", srv.TaskHandler)
	// Defines the /result path and uses ResultHandler to handle data
	http.HandleFunc("/result", srv.ResultHandler)

	// --------------------------------
	// CLI ENDPOINTS
	// --------------------------------
	// These all require localhost and must be ran on the server host

	// Defines the /get-agents path and returns a list of connected agents
	http.HandleFunc("/get-agents", requireLocalhost(srv.GetAgentsHandler))
	// Defines the /enqueue path and enqueues tasks to list of targeted agents
	http.HandleFunc("/enqueue", requireLocalhost(srv.EnqueueHandler))
	// Defines the /add-targets path and allows targeting of agents
	http.HandleFunc("/add-targets", requireLocalhost(srv.AddTargetsHandler))
	// Defines the /get-targets path and returns a list of targeted agents
	http.HandleFunc("/get-targets", requireLocalhost(srv.GetTargetsHandler))
	// Defines the /untarget path and untargets the provided agents
	http.HandleFunc("/untarget", requireLocalhost(srv.UntargetHandler))
	// Defines the /clear-targets path and clears all targets
	http.HandleFunc("/clear-targets", requireLocalhost(srv.ClearTargetsHandler))
	// Defines the /set-targets path and clears targets before adding the provided ones
	http.HandleFunc("/set-targets", requireLocalhost(srv.SetTargetsHandler))
	// Defines the /get-results path and sends results to the requester
	http.HandleFunc("/get-results", requireLocalhost(srv.GetResultsHandler))

	// Starts the server
	port := fmt.Sprintf(":%d", srv.port)
	err := http.ListenAndServe(port, nil)

	// Throws an error if the server fails to start
	if err != nil {
		return fmt.Errorf("Error: %w", err)
	}

	return nil
}
