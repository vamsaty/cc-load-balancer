package lb

import (
	"fmt"
	factory "github.com/vamsaty/cc-rate-limiter/limiter"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// LBConfig contains the configurations required for health checks
type LBConfig struct {
	// delay before load balancer can receive requests
	initialDelay time.Duration
	// interval at which health check occurs
	healthInterval time.Duration
	// time to wait before health check of an unhealthy server
	recheckInterval time.Duration
}

// LB a simple load balancer
type LB struct {
	rwLock   *sync.RWMutex   // to update the server's health
	wg       *sync.WaitGroup // wait for health check of all the servers
	config   *LBConfig       // healthCheck config
	servers  []*Endpoint     // list of endpoints
	CanServe bool            // load balancer can start serving
	index    int             // next server to forward request to

	Host string // host to listen on
	Port int    // port to listen on
	*zap.Logger

	// integrated custom rate limiter
	Limiter factory.RateLimiter
}

func (lb *LB) Address() string { return lb.Host + ":" + strconv.Itoa(lb.Port) }

// GetNextServer returns the next (healthy) server to forward the request to
func (lb *LB) GetNextServer() *Endpoint {
	// check index then update
	ans := lb.index
	for maxIter := len(lb.servers); maxIter > 0; maxIter-- {
		lb.index = (lb.index + 1) % len(lb.servers)
		if lb.servers[ans].IsHealthy {
			return lb.servers[ans]
		}
		ans = lb.index
	}
	lb.Warn("No healthy servers found")
	return nil
}

// HealthChecker checks the health of all servers, it uses a ticker to check
// the health of all servers periodically.
func (lb *LB) HealthChecker() {
	lb.healthCheck()
	lb.CanServe = true
	lb.Info("Load balancer can start serving")

	ticker := time.NewTicker(lb.config.healthInterval)
	done := make(chan bool)
	for {
		select {
		case <-ticker.C:
			lb.healthCheck()
		case <-done:
			return
		}
	}
}

// healthCheck checks the health of all servers, marks them as healthy/unhealthy.
func (lb *LB) healthCheck() {
	for _, server := range lb.servers {
		lb.wg.Add(1)
		go lb.updateHealth(server)
	}
	lb.wg.Wait() // wait for all the servers to be checked
}

// updateHealth checks the backend server's health and update the endpoint.
func (lb *LB) updateHealth(endpoint *Endpoint) {
	defer lb.wg.Done()

	ok := endpoint.HealthCheck()
	lb.rwLock.Lock()
	endpoint.IsHealthy = ok
	lb.rwLock.Unlock()

	lb.Info("info", zap.String(endpoint.Address(), endpoint.ShowStats()))
}

// ForwardRequest forwards the request to the next (healthy) server, if it exists.
// else response with 503 Service Unavailable.
func (lb *LB) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	// check if load balancer is ready to serve
	if !lb.CanServe {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Load Balancer is not ready yet"))
		return
	}

	fmt.Println("checking if request can be allowed for", r.Host)
	if lb.Limiter.Allow(r.Host) != nil {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too many requests"))
		return
	}

	server := lb.GetNextServer()
	if server == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("No servers available"))
		return
	}
	if remote, err := url.Parse(server.Address()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		server.IncRequestCount()

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = r.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = remote.Path
		}
		proxy.ServeHTTP(w, r)
	}
}

func (lb *LB) Start() {
	go lb.HealthChecker()

	log.Fatalf("Load Balancer failed. Error: %s",
		http.ListenAndServe(
			lb.Address(),
			http.HandlerFunc(lb.ForwardRequest),
		),
	)
}

// NewLB creates a new load balancer with the given servers.
// Uses default configurations
func NewLB(servers []*Endpoint, port int) *LB {
	logger, _ := zap.NewProduction()
	return &LB{
		rwLock:   &sync.RWMutex{},
		servers:  servers,
		CanServe: false,
		config: &LBConfig{
			initialDelay:    3 * time.Second,
			healthInterval:  5 * time.Second,
			recheckInterval: 2 * time.Second,
		},
		wg:     &sync.WaitGroup{},
		Logger: logger,
		index:  0,
		Port:   port,
		Limiter: factory.NewRateLimiterFromConfig(map[string]string{
			"algo": "dummy",
		}),
	}
}

func NewRateLimitedLB(servers []*Endpoint, port int) *LB {
	lb := NewLB(servers, port)
	lb.Limiter = factory.NewRateLimiterFromConfig(map[string]string{
		"algo":                "token_bucket",
		"bucket_capacity":     "10",
		"token_push_interval": "1s",
	})
	return lb
}

func NewRateLimitedLBWithConfig(servers []*Endpoint, port int, config map[string]string) *LB {
	lb := NewLB(servers, port)
	lb.Limiter = factory.NewRateLimiterFromConfig(config)
	return lb
}
