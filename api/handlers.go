package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/luizbafilho/fusis/types"
)

type ServiceResponse struct {
	types.Service
	Destinations []types.Destination
}

func (as ApiService) getServices(c echo.Context) error {
	svcs, err := as.balancer.GetServices()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, svcs)
}

func (as ApiService) getService(c echo.Context) error {
	service, err := as.balancer.GetService(c.Param("service_name"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, service)
}

func (as ApiService) addService(c echo.Context) error {
	var newService types.Service
	if err := c.Bind(&newService); err != nil {
		return err
	}

	if err := as.balancer.AddService(&newService); err != nil {
		return err
	}

	c.Response().Header().Add("Location", "/services/"+newService.Name)
	return c.JSON(http.StatusCreated, newService)
}

func (as ApiService) deleteService(c echo.Context) error {
	if err := as.balancer.DeleteService(c.Param("service_name")); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (as ApiService) addDestination(c echo.Context) error {
	svc, err := as.balancer.GetService(c.Param("service_name"))
	if err != nil {
		return err
	}

	dst := &types.Destination{ServiceId: svc.GetId()}
	if err := c.Bind(dst); err != nil {
		return err
	}

	if err := as.balancer.AddDestination(svc, dst); err != nil {
		return err
	}

	c.Response().Header().Add("Location", fmt.Sprintf("/services/%s/destinations/%s", svc.GetId(), dst.Name))
	return c.JSON(http.StatusCreated, dst)
}

func (as ApiService) deleteDestination(c echo.Context) error {
	if err := as.balancer.DeleteDestination(&types.Destination{Name: c.Param("destination_name")}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (as ApiService) addCheck(c echo.Context) error {
	var check types.CheckSpec
	if err := c.Bind(&check); err != nil {
		return err
	}

	check.ServiceID = c.Param("service_name")

	// Converting int to time.Duration
	check.Interval = check.Interval * time.Second
	check.Timeout = check.Timeout * time.Second

	if err := as.balancer.AddCheck(check); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, check)
}

func (as ApiService) deleteCheck(c echo.Context) error {
	svcId := c.Param("service_name")
	if _, err := as.balancer.GetService(svcId); err != nil {
		return err
	}

	if err := as.balancer.DeleteCheck(types.CheckSpec{ServiceID: svcId}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
