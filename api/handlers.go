package api

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/fusis/engine"
)

func (as ApiService) serviceList(c *gin.Context) {
	services, err := engine.GetServices()

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("GetServices() failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, *services)
}

func (as ApiService) serviceCreate(c *gin.Context) {
	var newService engine.Service

	if c.BindJSON(&newService) != nil {
		return
	}

	if _, errs := govalidator.ValidateStruct(newService); errs != nil {
		c.JSON(422, gin.H{"errors": govalidator.ErrorsByField(errs)})
		return
	}

	err := engine.AddService(&newService)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("UpsertService() failed: %v", err)})
	} else {
		c.JSON(http.StatusOK, newService)
	}
}

func (as ApiService) serviceDelete(c *gin.Context) {
	// serviceId := c.Param("service_id")
	// service, err := getServiceFromId(serviceId)
	//
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }
	//
	// err = as.store.DeleteService(*service)
	//
	// if err != nil {
	// 	c.JSON(422, gin.H{"error": fmt.Sprintf("DeleteService() failed: %v\n", err)})
	// } else {
	// 	c.Data(http.StatusOK, gin.MIMEHTML, nil)
	// }
}

func (as ApiService) destinationUpsert(c *gin.Context) {
	// serviceId := c.Param("service_id")
	// service, err := getServiceFromId(serviceId)
	//
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }
	//
	// destination := store.Destination{Weight: 1, Mode: "route"}
	//
	// if c.BindJSON(&destination) != nil {
	// 	return
	// }
	//
	// err = as.store.UpsertDestination(*service, destination)
	//
	// if err != nil {
	// 	c.JSON(422, gin.H{"error": fmt.Sprintf("UpsertDestination() failed: %v\n", err)})
	// } else {
	// 	c.JSON(http.StatusOK, destination)
	// }
}

func (as ApiService) destinationDelete(c *gin.Context) {
	// serviceId := c.Param("service_id")
	// destinationId := c.Param("destination_id")
	//
	// service, err := getServiceFromId(serviceId)
	// destination, err := getDestinationFromId(destinationId)
	//
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }
	//
	// err = as.store.DeleteDestination(*service, *destination)
	//
	// if err != nil {
	// 	c.JSON(422, gin.H{"error": fmt.Sprintf("DeleteDestination() failed: %v\n", err)})
	// } else {
	// 	c.Data(http.StatusOK, gin.MIMEHTML, nil)
	// }
}

func (as ApiService) flush(c *gin.Context) {
	// err := as.store.Flush()
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }
	//
	// err = ipvs.Flush()
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	return
	// }
}

// func getDestinationFromId(destinationId string) (*store.Destination, error) {
// 	destinationAttrs := strings.Split(destinationId, "-")
//
// 	port, err := strconv.ParseUint(destinationAttrs[1], 10, 16)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &store.Destination{
// 		Host: destinationAttrs[0],
// 		Port: uint16(port),
// 	}, nil
// }
