package be

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// BackendServer is the http server.
// The loadbalancer sends request to this server.
type BackendServer struct {
	Host string
	Port int
}

// Address returns the address of the server as host:port (socket)
func (s *BackendServer) Address() string { return s.Host + ":" + strconv.Itoa(s.Port) }

// Health returns the health check endpoint
func (s *BackendServer) Health() string { return s.Address() + "/health" }

func (s *BackendServer) Start() {
	fmt.Println("Starting BackendServer on", s.Address())

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "Health Check OK")
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello From Backend " + s.Address(),
		})
	})

	log.Fatalln(r.Run(s.Address()))
}

func NewBE(port int) *BackendServer {
	return &BackendServer{Port: port}
}
