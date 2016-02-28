package api

import (
	"fmt"
	"net/http"

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

	as.router.POST("/services", as.serviceCreate)
	// router.GET("/services", serviceIndex)
	// router.POST("/services", serviceCreate)
	// router.PUT("/services", serviceUpdate)
	// router.DELETE("/services", serviceDelete)
	//
	// router.POST("/services/:service_id/destinations", destinationCreate)
	// router.PUT("/services/:service_id/destinations", destinationUpdate)
	// router.DELETE("/services/:service_id/destinations", destinationDelete)

	as.router.Run(":8000")
}

func (as ApiService) Stop() {
	fmt.Println("Stoping gin...")
}

func (as ApiService) serviceCreate(c *gin.Context) {
	var newService store.ServiceRequest

	if c.BindJSON(&newService) != nil {
		return
	}

	err := as.store.AddService(newService)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.AddService() failed: %v", err)})
	} else {
		c.JSON(http.StatusCreated, newService)
	}
}
