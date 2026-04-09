package agent

import (
	"log"
	"net"
	"regexp"
	"time"
)

// GetLocalIP returns the local ip of the agent
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "?.?.?.?"
	}
	rePriority := []*regexp.Regexp{
		regexp.MustCompile(`^10\.(\d+)\.1\.10$`),
		regexp.MustCompile(`^10\.(\d+)\.1\.40$`),
		regexp.MustCompile(`^10\.(\d+)\.1\.30$`),
		regexp.MustCompile(`^10\.(\d+)\.1\.60$`),
		regexp.MustCompile(`^10\.(\d+)\.1\.70$`),
		regexp.MustCompile(`^10\.(\d+)\.1\.80$`),
		regexp.MustCompile(`^10\.(\d+)\.1\.90$`),
		regexp.MustCompile(`^10\.(\d+)\.2\.2$`),
		regexp.MustCompile(`^10\.(\d+)\.2\.4$`),
		regexp.MustCompile(`^10\.(\d+)\.2\.10$`),
		regexp.MustCompile(`^10\.(\d+)\.1\.1$`),
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				for _, re := range rePriority {
					if re.MatchString(ipStr) {
						return ipStr
					}
				}
			}
		}
	}
	return "?.?.?.?"
}

// registerUntilDone has the agent attempt to register with the server until it is accepted
func (a *Agent) registerUntilDone() {
	for {
		if err := a.comm.Register(a.Info, a.Debug); err != nil {
			if a.Debug {
				log.Println(err)
			}
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}
