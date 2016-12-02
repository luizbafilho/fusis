package fusis

import (
	"testing"
	"time"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OperationsTestSuite struct {
	suite.Suite

	balancer    Balancer
	service     types.Service
	destination types.Destination
}

func TestStateTestSuite(t *testing.T) {
	suite.Run(t, new(OperationsTestSuite))
}

func (s *OperationsTestSuite) SetupSuite() {
	var err error
	s.balancer, err = NewBalancer(defaultConfig())
	assert.Nil(s.T(), err)
	WaitForResult(func() (bool, error) {
		return s.balancer.IsLeader(), nil
	}, func(err error) {
		assert.Fail(s.T(), "balancer did not become leader")
	})

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
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *OperationsTestSuite) TearDownTest() {
	b := s.balancer.(*FusisBalancer)
	b.cleanup()
}

type testFn func() (bool, error)
type errorFn func(error)

func WaitForResult(test testFn, error errorFn) {
	retries := 1000

	for retries > 0 {
		time.Sleep(10 * time.Millisecond)
		retries--

		success, err := test()
		if success {
			return
		}

		if retries == 0 {
			error(err)
		}
	}
}

func (s *OperationsTestSuite) TestAddService() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(s.T(), []types.Service{s.service}, s.balancer.GetServices())

	//Asserting vip allocation
	assert.Equal(s.T(), "192.168.0.1", s.service.Address)

	//Asserting struct validation
	err = s.balancer.AddService(&types.Service{})
	errValidation := types.ErrValidation{
		Type: "service",
		Errors: map[string]string{
			"Port":      "field field must be greater than 1",
			"Protocol":  "field must be one of the following: tcp | udp",
			"Scheduler": "field must be one of the following: rr | wrr | lc",
			"Mode":      "field is required",
			"Name":      "field is required",
		},
	}
	assert.Equal(s.T(), errValidation, err)

	// Asserting conflict
	err = s.balancer.AddService(&s.service)
	assert.Equal(s.T(), types.ErrServiceConflict, err)
}

func (s *OperationsTestSuite) TestGetService() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	svc, err := s.balancer.GetService(s.service.Name)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.service, *svc)

	// Asserting service not found
	_, err = s.balancer.GetService("anything")
	assert.Equal(s.T(), err, types.ErrServiceNotFound)
}

func (s *OperationsTestSuite) TestDeleteService() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	err = s.balancer.DeleteService(s.service.Name)
	time.Sleep(500 * time.Millisecond)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), s.balancer.GetServices(), 0)

	// Asserting service not found
	err = s.balancer.DeleteService("anything")
	assert.Equal(s.T(), err, types.ErrServiceNotFound)
}

func (s *OperationsTestSuite) TestAddDestination() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)

	err = s.balancer.AddDestination(&s.service, &s.destination)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(s.T(), []types.Destination{s.destination}, s.balancer.GetDestinations(&s.service))

	//Asserting struct validation
	err = s.balancer.AddDestination(&s.service, &types.Destination{})
	errValidation := types.ErrValidation{
		Type: "destination",
		Errors: map[string]string{
			"Name":      "field is required",
			"Address":   "field is required",
			"Port":      "field field must be greater than 1",
			"ServiceId": "field is required",
		},
	}
	assert.Equal(s.T(), errValidation, err)

	// Asserting conflict
	err = s.balancer.AddDestination(&s.service, &s.destination)
	assert.Equal(s.T(), types.ErrDestinationConflict, err)

	//Asserting default
	dst := &types.Destination{}
	_ = s.balancer.AddDestination(&s.service, dst)
	assert.Equal(s.T(), "nat", dst.Mode)
	assert.Equal(s.T(), int32(1), dst.Weight)
}

func (s *OperationsTestSuite) TestDeleteDestination() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)

	err = s.balancer.AddDestination(&s.service, &s.destination)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	err = s.balancer.DeleteDestination(&s.destination)
	time.Sleep(500 * time.Millisecond)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), s.balancer.GetDestinations(&s.service), 0)

	// Asserting service not found
	err = s.balancer.DeleteDestination(&types.Destination{})
	assert.Equal(s.T(), err, types.ErrServiceNotFound)

	err = s.balancer.DeleteDestination(&types.Destination{ServiceId: s.service.GetId()})
	assert.Equal(s.T(), err, types.ErrDestinationNotFound)
}

func defaultConfig() *config.BalancerConfig {
	return &config.BalancerConfig{
		StoreAddress: "consul://localhost:8500",
		StorePrefix:  "fusis-dev",
		Interfaces: config.Interfaces{
			Inbound:  "lo",
			Outbound: "lo",
		},
		Name: "Test",
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/28"},
		},
	}
}
