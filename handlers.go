package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/janus/store"
	"github.com/luizbafilho/janus/store/etcd"
)

// //Index Handles index
// func serviceIndex(c *gin.Context) {
// 	ipvsServices, err := ipvs.GetServices()
//
// 	if err != nil {
// 		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.GetServices() failed: %v", err)})
// 		return
// 	}
//
// 	services := []ServiceRequest{}
//
// 	for _, s := range ipvsServices {
// 		services = append(services, newServiceRequest(s))
// 	}
//
// 	c.JSON(http.StatusOK, services)
// }
//
// //ServiceCreate ...
func serviceCreate(c *gin.Context) {
	var newService store.ServiceRequest

	if c.BindJSON(&newService) != nil {
		return
	}

	// err := ipvs.AddService(newService.toIpvsService())
	nodes := []string{"http://127.0.0.1:2379"}
	s := etcd.New(nodes)

	err := s.AddService(newService)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.AddService() failed: %v", err)})
	} else {
		c.JSON(http.StatusCreated, newService)
	}
}

//
// func serviceUpdate(c *gin.Context) {
// 	var service ServiceRequest
//
// 	if c.BindJSON(&service) != nil {
// 		return
// 	}
//
// 	ipvsService := service.toIpvsService()
//
// 	err := ipvs.UpdateService(ipvsService)
//
// 	if err != nil {
// 		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.UpdateService() failed: %v\n", err)})
// 	} else {
// 		for _, d := range ipvsService.Destinations {
// 			fmt.Println(d.Weight)
// 			err = ipvs.UpsertDestination(ipvsService, *d)
//
// 			if err != nil {
// 				c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.UpdateDestination() failed: %v\n", err)})
// 				return
// 			}
// 		}
// 		c.JSON(http.StatusOK, service)
// 	}
//
// }
//
// //
// func serviceDelete(c *gin.Context) {
// 	var service ServiceRequest
//
// 	if c.BindJSON(&service) != nil {
// 		return
// 	}
//
// 	err := ipvs.DeleteService(service.toIpvsService())
//
// 	if err != nil {
// 		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.DeleteService() failed: %v\n", err)})
// 	} else {
// 		c.JSON(http.StatusOK, service)
// 	}
// }
//
// func destinationCreate(c *gin.Context) {
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
// 	err = ipvs.AddDestination(*service, *destination.toIpvsDestination())
//
// 	if err != nil {
// 		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.AddDestination() failed: %v\n", err)})
// 	} else {
// 		c.JSON(http.StatusOK, destination)
// 	}
// }
//
// //
// func destinationUpdate(c *gin.Context) {
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
// 	err = ipvs.UpdateDestination(*service, *destination.toIpvsDestination())
//
// 	if err != nil {
// 		c.JSON(422, gin.H{"error": fmt.Sprintf("ipvs.UpdateDestination() failed: %v\n", err)})
// 	} else {
// 		c.JSON(http.StatusOK, destination)
// 	}
// }
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
// func getServiceFromId(serviceId string) (*ipvs.Service, error) {
// 	serviceAttrs := strings.Split(serviceId, "-")
//
// 	port, err := strconv.ParseUint(serviceAttrs[1], 10, 16)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var proto IPProto
// 	proto.UnmarshalJSON([]byte(serviceAttrs[2]))
//
// 	return &ipvs.Service{
// 		Address:  net.ParseIP(serviceAttrs[0]),
// 		Port:     uint16(port),
// 		Protocol: ipvs.IPProto(proto),
// 	}, nil
// }
