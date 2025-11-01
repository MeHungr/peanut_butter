// This contains the main server logic for peanut-butter
package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MeHungr/peanut-butter/internal/storage"
	"github.com/MeHungr/peanut-butter/internal/transport"
	srvtransport "github.com/MeHungr/peanut-butter/internal/transport/server"
)

type Server struct {
	storage *storage.Storage
	port    int
	comm    *srvtransport.CommManager
}

func New(storage *storage.Storage, port int) *Server {
	// CommManager
	cm := &srvtransport.CommManager{
		Storage:    storage,
		Transports: make(map[transport.TransportString]srvtransport.Transport),
		Agents:     make(map[string]srvtransport.Transport),
	}

	// Transports
	httpsTransport := &srvtransport.HTTPSTransport{
		Comm: cm,
	}

	// Add transports to map
	cm.Transports[transport.HTTPS] = httpsTransport

	// Return the server
	return &Server{
		storage: storage,
		port:    port,
		comm:    cm,
	}
}

// Start starts the server and starts listening on the specified port
func (srv *Server) Start() error {

	// --------------------------------
	// AGENT ENDPOINTS
	// --------------------------------
	srv.comm.Transports[transport.HTTPS].Start()

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
	log.Printf("Starting HTTPS server on %s\n", port)
	err := http.ListenAndServeTLS(port, "server.crt", "server.key", nil)

	// Throws an error if the server fails to start
	if err != nil {
		return fmt.Errorf("Error: %w", err)
	}

	return nil
}
