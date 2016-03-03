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
	as.router.POST("/services", as.serviceUpsert)
	as.router.DELETE("/services/:service_id", as.serviceDelete)

	as.router.POST("/services/:service_id/destinations", as.destinationUpsert)
	as.router.DELETE("/services/:service_id/destinations/:destination_id", as.destinationDelete)

	as.router.Run(":8000")
}
