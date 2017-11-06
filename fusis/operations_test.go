package fusis

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OperationsTestSuite struct {
	suite.Suite

	balancer    Balancer
	service     types.Service
	destination types.Destination

	kv *clientv3.Client
}

func TestOperationsSuite(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}
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

	kv, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"172.100.0.40:2379"},
		DialTimeout: 5 * time.Second,
	})
	assert.Nil(s.T(), err)
	s.kv = kv
}

func (s *OperationsTestSuite) SetupTest() {
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
	_, err := s.kv.Delete(context.TODO(), defaultConfig().StorePrefix, clientv3.WithPrefix())
	assert.Nil(s.T(), err)

	s.balancer.(*FusisBalancer).flushVips()

	if out, err := exec.Command("ipvsadm", "--clear").CombinedOutput(); err != nil {
		log.Fatal(fmt.Errorf(" with message: `%s`, error: %v", strings.TrimSpace(string(out)), err))
	}
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
	time.Sleep(2 * time.Second)

	svcs, err := s.balancer.GetServices()
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), []types.Service{s.service}, svcs)
}

func (s *OperationsTestSuite) TestAddService_VipAllocation() {
	s.service.Address = ""
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	//Asserting vip allocation
	assert.Equal(s.T(), "192.168.0.1", s.service.Address)
}

func (s *OperationsTestSuite) TestAddService_Validation() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

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
}

func (s *OperationsTestSuite) TestAddService_Uniqueness() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	//Asserting name uniqueness
	err = s.balancer.AddService(&s.service)
	errValidation := types.ErrValidation{
		Type: "service",
		Errors: map[string]string{
			"ipvs": "service must be unique",
		},
	}
	assert.Equal(s.T(), errValidation, err)

	//Asserting ipvs uniqueness
	s.service.Name = "different-name"
	err = s.balancer.AddService(&s.service)
	errValidation = types.ErrValidation{
		Type: "service",
		Errors: map[string]string{
			"ipvs": "service must be unique",
		},
	}
	assert.Equal(s.T(), errValidation, err)
}

func (s *OperationsTestSuite) TestGetService() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(2000 * time.Millisecond)

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
	time.Sleep(1000 * time.Millisecond)

	err = s.balancer.DeleteService(s.service.Name)
	time.Sleep(1000 * time.Millisecond)
	assert.Nil(s.T(), err)

	svcs, err := s.balancer.GetServices()
	assert.Nil(s.T(), err)

	assert.Len(s.T(), svcs, 0)

	// Asserting service not found
	err = s.balancer.DeleteService("anything")
	assert.Equal(s.T(), err, types.ErrServiceNotFound)
}

func (s *OperationsTestSuite) TestAddDestination() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)

	err = s.balancer.AddDestination(&s.service, &s.destination)
	assert.Nil(s.T(), err)
	time.Sleep(2000 * time.Millisecond)

	dsts, err := s.balancer.GetDestinations(&s.service)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), []types.Destination{s.destination}, dsts)

	//Asserting default
	dst := &types.Destination{}
	_ = s.balancer.AddDestination(&s.service, dst)
	assert.Equal(s.T(), "nat", dst.Mode)
	assert.Equal(s.T(), int32(1), dst.Weight)
}

func (s *OperationsTestSuite) TestAddDestination_SetDefaults() {
	//Asserting default
	dst := &types.Destination{}
	_ = s.balancer.AddDestination(&s.service, dst)
	assert.Equal(s.T(), "nat", dst.Mode)
	assert.Equal(s.T(), int32(1), dst.Weight)
}

func (s *OperationsTestSuite) TestAddDestination_Validation() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)

	err = s.balancer.AddDestination(&s.service, &s.destination)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

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
}

func (s *OperationsTestSuite) TestAddDestination_Uniqueness() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)

	err = s.balancer.AddDestination(&s.service, &s.destination)
	assert.Nil(s.T(), err)
	time.Sleep(500 * time.Millisecond)

	//Asserting name uniqueness
	err = s.balancer.AddDestination(&s.service, &s.destination)
	errValidation := types.ErrValidation{
		Type: "destination",
		Errors: map[string]string{
			"ipvs": "destination must be unique",
		},
	}
	assert.Equal(s.T(), errValidation, err)

	//Asserting ivpvs uniqueness
	s.destination.Name = "unique-name"
	err = s.balancer.AddDestination(&s.service, &s.destination)
	errValidation = types.ErrValidation{
		Type: "destination",
		Errors: map[string]string{
			"ipvs": "destination must be unique",
		},
	}
	assert.Equal(s.T(), errValidation, err)
}

func (s *OperationsTestSuite) TestDeleteDestination() {
	err := s.balancer.AddService(&s.service)
	assert.Nil(s.T(), err)
	time.Sleep(1 * time.Second)

	err = s.balancer.AddDestination(&s.service, &s.destination)
	assert.Nil(s.T(), err)
	time.Sleep(1 * time.Second)

	err = s.balancer.DeleteDestination(&s.destination)
	time.Sleep(1 * time.Second)
	assert.Nil(s.T(), err)

	dsts, err := s.balancer.GetDestinations(&s.service)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), dsts, 0)

	// Asserting service not found
	err = s.balancer.DeleteDestination(&types.Destination{})
	assert.Equal(s.T(), err, types.ErrServiceNotFound)

	err = s.balancer.DeleteDestination(&types.Destination{ServiceId: s.service.GetId()})
	assert.Equal(s.T(), err, types.ErrDestinationNotFound)
}

func defaultConfig() *config.BalancerConfig {
	address := os.Getenv("ETCD_ENDPOINTS")
	if address == "" {
		address = "consul://localhost:8500"
	}
	return &config.BalancerConfig{
		EtcdEndpoints: address,
		StorePrefix:   "fusis-dev",
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
