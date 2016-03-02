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

func (as ApiService) serviceCreate(c *gin.Context) {
	var newService store.ServiceRequest

	if c.BindJSON(&newService) != nil {
		return
	}

	err := as.store.AddService(newService)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("AddService() failed: %v", err)})
	} else {
		c.JSON(http.StatusCreated, newService)
	}
}

func (as ApiService) serviceUpdate(c *gin.Context) {
	var service store.ServiceRequest

	if c.BindJSON(&service) != nil {
		return
	}

	err := as.store.UpdateService(service)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("UpdateService() failed: %v", err)})
	} else {
		c.JSON(http.StatusOK, service)
	}
}

func (as ApiService) serviceDelete(c *gin.Context) {
	var service store.ServiceRequest

	if c.BindJSON(&service) != nil {
		return
	}

	err := as.store.DeleteService(service)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("DeleteService() failed: %v\n", err)})
	} else {
		c.JSON(http.StatusOK, service)
	}
}

func (as ApiService) destinationCreate(c *gin.Context) {
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

	err = as.store.AddDestination(*service, destination)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("AddDestination() failed: %v\n", err)})
	} else {
		c.JSON(http.StatusOK, destination)
	}
}

func (as ApiService) destinationUpdate(c *gin.Context) {
	serviceId := c.Param("service_id")
	service, err := getServiceFromId(serviceId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var destination store.DestinationRequest

	if c.BindJSON(&destination) != nil {
		return
	}

	// err = ipvs.UpdateDestination(*service, *destination.toIpvsDestination())
	err = as.store.UpdateDestination(*service, destination)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("UpdateDestination() failed: %v\n", err)})
	} else {
		c.JSON(http.StatusOK, destination)
	}
}

//
// //
// func destinationDelete(c *gin.Context) {
// 	serviceId := c.Param("service_id")
// 	service, err := getServiceFromId(serviceId)
//
// 	if err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}
//
// 	var destination DestinationRequest
//
// 	if c.BindJSON(&destination) != nil {
// 		return
// 	}
//
// 	err = ipvs.DeleteDestination(*service, *destination.toIpvsDestination())
//
// 	if err != nil {
// 		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.DeleteDestination() failed: %v\n", err)})
// 	} else {
// 		c.JSON(http.StatusOK, destination)
// 	}
// }
//
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
