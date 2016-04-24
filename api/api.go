package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/fusis/fusis"
)

// ApiService ...
type ApiService struct {
	balancer *fusis.Balancer
	router   *gin.Engine
	env      string
}

//NewAPI ...
func NewAPI(balancer *fusis.Balancer) ApiService {
	gin.SetMode(gin.ReleaseMode)

	return ApiService{
		balancer: balancer,
		router:   gin.Default(),
		env:      getEnv(),
	}
}

func (as ApiService) Serve() {
	as.router.GET("/services", as.serviceList)
	as.router.GET("/services/:service_id", as.serviceGet)
	as.router.POST("/services", as.serviceCreate)
	as.router.DELETE("/services/:service_id", as.serviceDelete)

	as.router.POST("/services/:service_id/destinations", as.destinationCreate)
	as.router.DELETE("/services/:service_id/destinations/:destination_id", as.destinationDelete)

	if as.env == "test" {
		as.router.POST("/flush", as.flush)
	}
	as.router.Run("0.0.0.0:8000")
}

func getEnv() string {
	env := os.Getenv("FUSIS_ENV")
	if env == "" {
		env = "development"
	}
	return env
}
