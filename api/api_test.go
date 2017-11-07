package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/appleboy/gofight"
	"github.com/fatih/structs"
	"github.com/luizbafilho/fusis/fusis/mocks"
	"github.com/luizbafilho/fusis/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ApiTestSuite struct {
	suite.Suite

	api      ApiService
	balancer *mocks.Balancer

	r           *gofight.RequestConfig
	service     types.Service
	destination types.Destination
}

func TestStateTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func (s *ApiTestSuite) SetupTest() {
	s.balancer = new(mocks.Balancer)
	s.api = NewAPI(s.balancer)

	s.r = gofight.New()
	s.service = types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	s.destination = types.Destination{
		Name:      "test",
		Address:   "192.168.1.1",
		Port:      80,
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *ApiTestSuite) TestGetServices() {
	s.balancer.On("GetServices").Return([]types.Service{}, nil)

	expectedBody, err := json.Marshal([]types.Service{})
	assert.Nil(s.T(), err)

	s.r.GET("/services").
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestGetService() {
	s.balancer.On("GetService", "testing").Return(&s.service, nil)
	s.balancer.On("GetDestinations", &s.service).Return([]types.Destination{s.destination}, nil)

	resp := structs.Map(s.service)
	resp["Destinations"] = []types.Destination{s.destination}
	expectedBody, err := json.Marshal(resp)

	assert.Nil(s.T(), err)

	s.r.GET("/services/testing").
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestGetService_NotFound() {
	s.balancer.On("GetService", "testing").Return(&s.service, types.ErrServiceNotFound)

	expectedBody, _ := json.Marshal(map[string]string{
		"error": types.ErrServiceNotFound.Error(),
	})

	s.r.GET("/services/testing").
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(s.T(), http.StatusNotFound, r.Code)
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestAddService() {
	s.balancer.On("AddService", &s.service).Return(nil)

	expectedBody, err := json.Marshal(s.service)
	assert.Nil(s.T(), err)

	s.r.POST("/services").
		SetJSON(gofight.D{
			"name":      s.service.Name,
			"address":   s.service.Address,
			"port":      s.service.Port,
			"mode":      s.service.Mode,
			"scheduler": s.service.Scheduler,
			"protocol":  s.service.Protocol,
		}).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusCreated, r.Code)
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestAddService_Validation() {
	err := types.ErrValidation{
		Type: "service",
		Errors: map[string]string{
			"name": "field is required",
		},
	}
	s.balancer.On("AddService", &types.Service{Address: s.service.Address}).Return(err)

	expectedBody, _ := json.Marshal(map[string]interface{}{
		"error": err.Errors,
	})

	s.r.POST("/services").
		SetJSON(gofight.D{
			"address": s.service.Address,
		}).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusUnprocessableEntity, r.Code)
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestAddService_Conflict() {
	s.balancer.On("AddService", &s.service).Return(types.ErrServiceConflict)

	expectedBody, _ := json.Marshal(map[string]interface{}{
		"error": types.ErrServiceConflict.Error(),
	})

	s.r.POST("/services").
		SetJSON(gofight.D{
			"name":      s.service.Name,
			"address":   s.service.Address,
			"port":      s.service.Port,
			"mode":      s.service.Mode,
			"scheduler": s.service.Scheduler,
			"protocol":  s.service.Protocol,
		}).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusConflict, r.Code)
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestDeleteService() {
	s.balancer.On("DeleteService", s.service.Name).Return(nil)

	s.r.DELETE("/services/"+s.service.Name).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusNoContent, r.Code)
			assert.Equal(s.T(), "", r.Body.String())
		})
}

func (s *ApiTestSuite) TestDeleteService_NotFound() {
	s.balancer.On("DeleteService", s.service.Name).Return(types.ErrServiceNotFound)

	expectedBody, _ := json.Marshal(map[string]interface{}{
		"error": types.ErrServiceNotFound.Error(),
	})

	s.r.DELETE("/services/"+s.service.Name).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusNotFound, r.Code)
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestAddDestination() {
	s.balancer.On("GetService", s.service.Name).Return(&s.service, nil)
	s.balancer.On("AddDestination", &s.service, &s.destination).Return(nil)

	expectedBody, err := json.Marshal(s.destination)
	assert.Nil(s.T(), err)

	s.r.POST("/services/"+s.service.Name+"/destinations").
		SetJSON(gofight.D{
			"name":    s.destination.Name,
			"address": s.destination.Address,
			"weight":  1,
			"port":    s.destination.Port,
		}).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusCreated, r.Code)
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestDeleteDestination() {
	s.balancer.On("DeleteDestination", &types.Destination{Name: s.destination.Name}).Return(nil)

	path := fmt.Sprintf("/services/%s/destinations/%s", s.service.Name, s.destination.Name)
	s.r.DELETE(path).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusNoContent, r.Code)
			assert.Equal(s.T(), "", r.Body.String())
		})
}

func (s *ApiTestSuite) TestAddCheck() {
	spec := types.CheckSpec{ServiceID: s.service.Name, Interval: 5 * time.Second, Timeout: 5 * time.Second}
	s.balancer.On("AddCheck", spec).Return(nil)

	spec.ServiceID = s.service.GetId()
	expectedBody, err := json.Marshal(spec)
	assert.Nil(s.T(), err)

	s.r.POST("/services/"+s.service.Name+"/check").
		SetJSON(gofight.D{
			"interval": 5,
			"timeout":  5,
		}).
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusCreated, r.Code)
			assert.Equal(s.T(), string(expectedBody), r.Body.String())
		})
}

func (s *ApiTestSuite) TestAddDelete() {
	s.balancer.On("GetService", s.service.GetId()).Return(&s.service, nil)

	spec := types.CheckSpec{ServiceID: s.service.Name}
	s.balancer.On("DeleteCheck", spec).Return(nil)

	s.r.DELETE("/services/"+s.service.Name+"/check").
		Run(s.api.echo, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			s.balancer.AssertExpectations(s.T())

			assert.Equal(s.T(), http.StatusNoContent, r.Code)
			assert.Equal(s.T(), "", r.Body.String())
		})
}
