package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/fusis/ipvs"
)

// ApiService ...
type ApiService struct {
	balancer Balancer
	router   *gin.Engine
	env      string
}

type Balancer interface {
	GetServices() []ipvs.Service
	AddService(*ipvs.Service) error
	GetService(string) (*ipvs.Service, error)
	DeleteService(string) error
	AddDestination(*ipvs.Service, *ipvs.Destination) error
	GetDestination(string) (*ipvs.Destination, error)
	DeleteDestination(*ipvs.Destination) error
}

//NewAPI ...
func NewAPI(balancer Balancer) ApiService {
	gin.SetMode(gin.ReleaseMode)
	as := ApiService{
		balancer: balancer,
		router:   gin.Default(),
		env:      getEnv(),
	}
	as.registerRoutes()
	return as
}

func (as ApiService) registerRoutes() {
	as.router.GET("/services", as.serviceList)
	as.router.GET("/services/:service_id", as.serviceGet)
	as.router.POST("/services", as.serviceCreate)
	as.router.DELETE("/services/:service_id", as.serviceDelete)
	as.router.POST("/services/:service_id/destinations", as.destinationCreate)
	as.router.DELETE("/services/:service_id/destinations/:destination_id", as.destinationDelete)
	if as.env == "test" {
		as.router.POST("/flush", as.flush)
	}
}

func (as ApiService) Serve() {
	as.router.Run("0.0.0.0:8000")
}

func getEnv() string {
	env := os.Getenv("FUSIS_ENV")
	if env == "" {
		env = "development"
	}
	return env
}
