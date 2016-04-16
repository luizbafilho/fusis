package api

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/asdine/storm"
	"github.com/gin-gonic/gin"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/engine/store"
)

func (as ApiService) serviceList(c *gin.Context) {
	services, err := engine.GetServices()

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("GetServices() failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, services)
}

func (as ApiService) serviceGet(c *gin.Context) {
	serviceId := c.Param("service_id")
	service, err := engine.GetService(serviceId)

	if err != nil {
		if err == storm.ErrNotFound {
			c.JSON(404, gin.H{"error": fmt.Sprintf("GetService(): %v", err)})
		} else {
			c.JSON(422, gin.H{"error": fmt.Sprintf("GetService() failed: %v", err)})
		}
		return
	}

	c.JSON(http.StatusOK, service)
}

func (as ApiService) serviceCreate(c *gin.Context) {
	newService := store.Service{}

	if c.BindJSON(&newService) != nil {
		return
	}
	//Guarantees that no one tries to create a destination together with a service
	newService.Destinations = []store.Destination{}

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
	serviceId := c.Param("service_id")
	err := engine.DeleteService(serviceId)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("DeleteService() failed: %v\n", err)})
	} else {
		c.Data(http.StatusOK, gin.MIMEHTML, nil)
	}
}

func (as ApiService) destinationCreate(c *gin.Context) {
	serviceId := c.Param("service_id")
	service, err := engine.GetService(serviceId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	destination := &store.Destination{Weight: 1, Mode: "route", ServiceId: serviceId}

	if c.BindJSON(destination) != nil {
		return
	}

	err = engine.AddDestination(service, destination)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("UpsertDestination() failed: %v\n", err)})
	} else {
		c.JSON(http.StatusOK, destination)
	}
}

func (as ApiService) destinationDelete(c *gin.Context) {
	destinationId := c.Param("destination_id")

	err := engine.DeleteDestination(destinationId)

	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("DeleteDestination() failed: %v\n", err)})
	} else {
		c.Data(http.StatusOK, gin.MIMEHTML, nil)
	}
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
