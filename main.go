package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/janus/ipvs"
)

func main() {
	if err := ipvs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v\n", err)
	}
	log.Printf("IPVS version %s\n", ipvs.Version())

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()

	router.GET("/services", serviceIndex)
	router.POST("/services", serviceCreate)
	router.PUT("/services", serviceUpdate)
	router.DELETE("/services", serviceDelete)
	//
	router.POST("/services/:service_id/destinations", destinationCreate)
	router.PUT("/services/:service_id/destinations", destinationUpdate)
	router.DELETE("/services/:service_id/destinations", destinationDelete)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run(":8000")
	// router.Run(":3000") for a hard coded port
}
