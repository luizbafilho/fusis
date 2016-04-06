package api

import "github.com/gin-gonic/gin"

type ApiService struct {
	router *gin.Engine
	env    string
}

func NewAPI(env string) ApiService {
	return ApiService{
		router: gin.Default(),
		env:    env,
	}
}

func (as ApiService) Serve() {
	as.router.GET("/services", as.serviceList)
	as.router.POST("/services", as.serviceCreate)
	as.router.DELETE("/services/:service_id", as.serviceDelete)

	as.router.POST("/services/:service_id/destinations", as.destinationCreate)
	as.router.DELETE("/services/:service_id/destinations/:destination_id", as.destinationDelete)

	if as.env == "test" {
		as.router.POST("/flush", as.flush)
	}

	as.router.Run(":8000")
}
