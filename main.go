package main

import (
	"flag"
	"github.com/vamsaty/cc-load-balancer/be"
	"github.com/vamsaty/cc-load-balancer/lb"
	"strconv"
	"strings"
)

var (
	isLoadBalancer = flag.Bool("lb", false, "Indicates if this binary runs as a load balancer or backend server")
	endpoints      = flag.String("lb-endpoints", "", "comma separated list of endpoints. eg: localhost:8080,localhost:8081")
	port           = flag.Int("port", 80, "load balancer port")
)

// GetHostPort returns the host and port from the string
func GetHostPort(s string) (string, int) {
	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return parts[0], port
		}
	}
	return "localhost", 8080 // default
}

// ParseEndpoints parses the comma separated list of endpoints into host:port pairs
func ParseEndpoints() []*lb.Endpoint {
	endPoints := strings.Split(*endpoints, ",")
	if len(endPoints) == 0 {
		panic("no endpoints provided")
	}

	var servers []*lb.Endpoint
	for _, ep := range endPoints {
		host, port := GetHostPort(ep)
		servers = append(servers, lb.NewEndpoint(host, port))
	}
	return servers
}

func main() {
	flag.Parse()
	if *isLoadBalancer {
		servers := ParseEndpoints()
		lb.NewRateLimitedLB(servers, *port).Start()
	} else {
		be.NewBE(*port).Start()
	}
}
