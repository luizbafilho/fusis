package api

import (
	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/janus/store"
)

type ApiService struct {
	store  store.Store
	router *gin.Engine
}

func NewAPI(store store.Store) ApiService {
	return ApiService{
		store:  store,
		router: gin.Default(),
	}
}

func (as ApiService) Serve() {

	as.router.GET("/services", as.serviceList)
	as.router.POST("/services", as.serviceCreate)
	as.router.PUT("/services", as.serviceUpdate)
	as.router.DELETE("/services", as.serviceDelete)

	as.router.POST("/services/:service_id/destinations", as.destinationCreate)
	// router.PUT("/services/:service_id/destinations", destinationUpdate)
	// router.DELETE("/services/:service_id/destinations", destinationDelete)

	as.router.Run(":8000")
}
