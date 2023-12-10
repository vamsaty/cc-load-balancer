package lb

import (
	"fmt"
	"net/http"
	"time"
)

// stats stores the metrics for an Endpoint
type stats struct {
	startTime      time.Time
	requestCount   int
	healthyCount   int
	unhealthyCount int
}

func (s *stats) IncRequestCount()   { s.requestCount++ }
func (s *stats) IncHealthyCount()   { s.healthyCount++ }
func (s *stats) IncUnhealthyCount() { s.unhealthyCount++ }

// Endpoint represents a backend server for Loadbalancer,
// responsible for making health checks
type Endpoint struct {
	Host     string
	Protocol string

	Port            int
	IsHealthy       bool
	LastHealthCheck time.Time
	stats
	http.Client
}

// ShowStats returns a string representation of the stats
func (ep *Endpoint) ShowStats() string {
	return fmt.Sprintf(
		"#req: %d, #good: %d, #bad: %d",
		ep.requestCount, ep.healthyCount, ep.unhealthyCount,
	)
}

func (ep *Endpoint) Address() string {
	return fmt.Sprintf("%s://%s:%d", ep.Protocol, ep.Host, ep.Port)
}

func (ep *Endpoint) Health() string { return ep.Address() + "/health" }

// HealthCheck makes a request to the server and updates the stats
func (ep *Endpoint) HealthCheck() bool {
	ep.LastHealthCheck = time.Now()
	if response, err := ep.Client.Get(ep.Health()); err == nil { // ok request
		if response.StatusCode == http.StatusOK {
			ep.IncHealthyCount()
			return true
		}
	}
	ep.IncUnhealthyCount() // update metrics
	return false
}

func NewEndpoint(host string, port int) *Endpoint {
	// indicates that the server has never been checked
	return &Endpoint{
		Host: host,
		Client: http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:  false,
				DisableCompression: false,
				MaxIdleConns:       1,
				IdleConnTimeout:    10,
			},
			Timeout: 5 * time.Second,
		},
		Protocol:        "http",
		Port:            port,
		IsHealthy:       false,
		LastHealthCheck: time.Now().AddDate(-100, 0, 0),
		stats:           stats{startTime: time.Now()},
	}
}
