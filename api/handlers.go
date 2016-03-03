package api

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/janus/ipvs"
	"github.com/luizbafilho/janus/store"
)

func (as ApiService) serviceList(c *gin.Context) {
	ipvsServices, err := ipvs.GetServices()

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.GetServices() failed: %v", err)})
		return
	}

	services := []store.ServiceRequest{}

	for _, s := range ipvsServices {
		services = append(services, store.NewServiceRequest(s))
	}

	c.JSON(http.StatusOK, services)
}

func (as ApiService) serviceUpsert(c *gin.Context) {
	var newService store.ServiceRequest

	if c.BindJSON(&newService) != nil {
		return
	}

	err := as.store.UpsertService(newService)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("UpsertService() failed: %v", err)})
	} else {
		c.JSON(http.StatusOK, newService)
	}
}

func (as ApiService) serviceDelete(c *gin.Context) {
	serviceId := c.Param("service_id")
	service, err := getServiceFromId(serviceId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = as.store.DeleteService(*service)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("DeleteService() failed: %v\n", err)})
	} else {
		c.Data(http.StatusOK, gin.MIMEHTML, nil)
	}
}

func (as ApiService) destinationUpsert(c *gin.Context) {
	serviceId := c.Param("service_id")
	service, err := getServiceFromId(serviceId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	destination := store.DestinationRequest{Weight: 1, Mode: store.RouteMode}

	if c.BindJSON(&destination) != nil {
		return
	}

	err = as.store.UpsertDestination(*service, destination)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("UpsertDestination() failed: %v\n", err)})
	} else {
		c.JSON(http.StatusOK, destination)
	}
}

func (as ApiService) destinationDelete(c *gin.Context) {
	serviceId := c.Param("service_id")
	destinationId := c.Param("destination_id")

	service, err := getServiceFromId(serviceId)
	destination, err := getDestinationFromId(destinationId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = as.store.DeleteDestination(*service, *destination)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("DeleteDestination() failed: %v\n", err)})
	} else {
		c.Data(http.StatusOK, gin.MIMEHTML, nil)
	}
}

func getServiceFromId(serviceId string) (*store.ServiceRequest, error) {
	serviceAttrs := strings.Split(serviceId, "-")

	port, err := strconv.ParseUint(serviceAttrs[1], 10, 16)

	if err != nil {
		return nil, err
	}

	var proto store.IPProto
	proto.UnmarshalJSON([]byte(serviceAttrs[2]))

	return &store.ServiceRequest{
		Host:     net.ParseIP(serviceAttrs[0]),
		Port:     uint16(port),
		Protocol: store.IPProto(proto),
	}, nil
}

func getDestinationFromId(destinationId string) (*store.DestinationRequest, error) {
	destinationAttrs := strings.Split(destinationId, "-")

	port, err := strconv.ParseUint(destinationAttrs[1], 10, 16)

	if err != nil {
		return nil, err
	}

	return &store.DestinationRequest{
		Host: net.ParseIP(destinationAttrs[0]),
		Port: uint16(port),
	}, nil
}
