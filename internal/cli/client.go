package cli

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"
)

// CACertPEM is the certificate from the CA used by the server
var CACertPEM = []byte(`
-----BEGIN CERTIFICATE-----
MIIFAzCCAuugAwIBAgIUKNb0ica95BeND8E3Gyy7iTLwA64wDQYJKoZIhvcNAQEL
BQAwETEPMA0GA1UEAwwGUE5VVEJSMB4XDTI1MTEwMTAzNDkzNFoXDTM1MTAzMDAz
NDkzNFowETEPMA0GA1UEAwwGUE5VVEJSMIICIjANBgkqhkiG9w0BAQEFAAOCAg8A
MIICCgKCAgEAwv1FCu24muJk/Vv88a/tT/IuZdHssylIx3HS3tiyPehVUmgumwK+
9Gk2J/E1Xad+u6CmeR7+nueyarVizDF38Z482U7ZnN8eKIi02YDrP5TAb9Wj374M
Axaxbtrmga8EH9lw5OoI9qvA/YmERVW/d5HxgqnfRWcvpvwRwRn70s2k/G3FKs7N
8kw8l9yI4cJcJJ1T3N95DaFkwQDVml43/saIlavwGZ4xexjkHoN8Z+MqOqs9IWsc
oDdcdp0TZlPXiseS7O6g/k99IsMDhFlK6KvFImYV3MtpAVoILR207udwUthWviVt
ZzjEoRi+AfjkiJxIZDrHY9CIYQ5jXAyw5XM8y64H15q+il3u6ZckJE65pLiGNxDg
nEpBxLJuSY0KoerDkddIpcRtsm+fSDGEUOQSmubPwCNlYKgIyBsRMAIYF8StjTXc
d5iSlxxpHV0HGMNIO5J/4iCDJAyEQ3QudUKe+3sJOEJH200+Evkma8UAjZGSBXJV
sWwTqRKFxit0sODAiUIExoLslVENkvJsT8KcHDQW5OhLnlEqqi622MFSm3e0mPwn
jnSfcVsY3umgGVZGQ9ofE9AQG5BGyVOZ/3zKLJMv29Ao4USduHgx079NkcepZ23S
Ua3N1Dnn/s0oqcnTlvoBsOOJDhd5c6nFOrMEqweIOC6Aw8vaOPYdmvMCAwEAAaNT
MFEwHQYDVR0OBBYEFPwlHGbupKQf3JOGzv38vEzTJG8QMB8GA1UdIwQYMBaAFPwl
HGbupKQf3JOGzv38vEzTJG8QMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
BQADggIBAEc9N4nNJXqCc6cS31GG076mE86pRqY0pifUj25qG9AF2DzcvH4N6VBI
ZvHpQSjGdgoz6PVe8cBWlkosjj5fSUs6pjNBhwQH8B1D0d6VvBo7sLaArBZqGpKK
0V8mG6ReaM3ZM+sYJf0h+KJL/lIG/9OgFdAzCTnOI+P9uTWc2Al9ZJFn4x0v1NY+
AZiBNVJxdFpYlIqQEeygX0YcGcnSWaE0oq3CD+bXowoAhXk+FPLNMoccchF1syOg
FBZtBpdCWDEbfm7LgpAULBMweu5LccyDvbUunDwOUL1e7ejdmrZMkU+sxm9E6330
F8w8jOdQRYBoJKSJeUhZt/diwgU35MiV3J1/Yto3eQ0s9liKMInbsZmHYvmMVlEE
WCfKAFX+biuCycs0NRvgMp5j/Ae3ZRTqxKn0ii31QPqn/br15CYm9GdrIY1J4lHx
CXtm/Do79JT1i9d3sLIdJYeWvFH1JazjJUnaBz+uXgV+xtVqgXpYAV9JMn+srrhl
STH/2+CPPuOjOCnMDZbcjJl8whFtRcseB12a/ftJpc/7u18gwAkhjhEAa8pGo44D
I93MI2l2JpXtAu1MIAXRXrM/oSe3D64b/iF4HSxrsUq8sWnZilaz0fTDVKjN7xKw
1IHvSFFAKQ7GMCYBN/W+tE4KYMwn6lwe9Aa6j/ls+s2SIYtcjjoN
-----END CERTIFICATE-----
`)

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

// Represents the CLI client
type Client struct {
	HTTPSClient *http.Client
	BaseURL     string
}

// NewCLIClient returns a CLI client with default values
func NewCLIClient(baseURL string) *Client {
	return &Client{
		HTTPSClient: NewHTTPSClient(),
		BaseURL:     baseURL,
	}
}
