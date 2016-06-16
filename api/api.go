package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/fusis/api/types"
)

// ApiService ...
type ApiService struct {
	balancer Balancer
	router   *gin.Engine
	env      string
}

type Balancer interface {
	GetServices() []types.Service
	AddService(*types.Service) error
	GetService(string) (*types.Service, error)
	DeleteService(string) error
	AddDestination(*types.Service, *types.Destination) error
	GetDestination(string) (*types.Destination, error)
	DeleteDestination(*types.Destination) error
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
	as.router.GET("/services/:service_name", as.serviceGet)
	as.router.POST("/services", as.serviceCreate)
	as.router.DELETE("/services/:service_name", as.serviceDelete)
	as.router.POST("/services/:service_name/destinations", as.destinationCreate)
	as.router.DELETE("/services/:service_name/destinations/:destination_name", as.destinationDelete)
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
