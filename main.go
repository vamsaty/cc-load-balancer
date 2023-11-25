package main

import (
	"flag"
	"github.com/vamsaty/cc-load-balancer/be"
	"github.com/vamsaty/cc-load-balancer/lb"
	"strconv"
	"strings"
)

var (
	isLoadBalancer = flag.Bool("lb", false, "is load balancer")
	endpoints      = flag.String("lb-endpoints", "", "load balancer port")
	port           = flag.Int("port", 80, "load balancer port")
)

func GetHostPort(s string) (string, int) {
	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return parts[0], port
		}
	}
	return "localhost", 8080 // default
}

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
		lb.NewLB(servers, *port).Start()
	} else {
		be.NewBE(*port).Start()
	}
}
