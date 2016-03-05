package etcd_test

import (
	"testing"

	. "github.com/luizbafilho/janus/store"
	"github.com/luizbafilho/janus/store/etcd"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type EtcdSuite struct {
	store       Store
	service     Service
	destination Destination
}

var _ = Suite(&EtcdSuite{})

func (s *EtcdSuite) SetUpSuite(c *C) {
	nodes := []string{"http://127.0.0.1:2379"}
	s.store = etcd.New(nodes, "fusis-test")

	s.service = Service{
		Host:      "10.7.0.1",
		Port:      80,
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	s.destination = Destination{
		Host:   "192.168.1.1",
		Port:   80,
		Mode:   "nat",
		Weight: 1,
	}
}

func (s *EtcdSuite) SetUpTest(c *C) {
	err := s.store.UpsertService(s.service)
	c.Assert(err, IsNil)
}

func (s *EtcdSuite) TearDownTest(c *C) {
	err := s.store.Flush()
	if err != nil {
		panic(err)
	}
}

func (s *EtcdSuite) TestGetService(c *C) {
	err := s.store.UpsertDestination(s.service, s.destination)
	c.Assert(err, IsNil)

	svc, err := s.store.GetService(s.service.GetId())
	c.Assert(err, IsNil)

	s.service.Destinations = []Destination{s.destination}

	c.Assert(*svc, DeepEquals, s.service)
}

func (s *EtcdSuite) TestGetServices(c *C) {
	svcs, err := s.store.GetServices()

	expected := []Service{s.service}

	c.Assert(err, IsNil)
	c.Assert(*svcs, DeepEquals, expected)
}

func (s *EtcdSuite) TestUpsertServices(c *C) {
	svc := Service{
		Host:      "10.8.0.1",
		Port:      8080,
		Scheduler: "rr",
		Protocol:  "tcp",
	}

	//Inserting service
	err := s.store.UpsertService(svc)
	c.Assert(err, IsNil)

	storedSvc, err := s.store.GetService(svc.GetId())
	c.Assert(err, IsNil)

	c.Assert(*storedSvc, DeepEquals, svc)

	//Updating service
	svc.Scheduler = "rr"
	err = s.store.UpsertService(svc)
	c.Assert(err, IsNil)

	storedSvc, err = s.store.GetService(svc.GetId())
	c.Assert(err, IsNil)

	c.Assert(*storedSvc, DeepEquals, svc)
}

func (s *EtcdSuite) TestDeleteService(c *C) {
	err := s.store.DeleteService(s.service)
	c.Assert(err, IsNil)

	_, err = s.store.GetService(s.service.GetId())
	c.Assert(err, ErrorMatches, ".*Key not found.*")
}

func (s *EtcdSuite) TestUpsertDestination(c *C) {
	//Inserting destination
	err := s.store.UpsertDestination(s.service, s.destination)
	c.Assert(err, IsNil)

	storedDsts, err := s.store.GetDestinations(s.service)
	c.Assert(err, IsNil)

	c.Assert(*storedDsts, DeepEquals, []Destination{s.destination})

	// Updating service
	s.destination.Weight = 2
	err = s.store.UpsertDestination(s.service, s.destination)
	c.Assert(err, IsNil)

	storedDsts, err = s.store.GetDestinations(s.service)
	c.Assert(err, IsNil)

	c.Assert(*storedDsts, DeepEquals, []Destination{s.destination})
}

func (s *EtcdSuite) TestDeleteDestination(c *C) {
	err := s.store.UpsertDestination(s.service, s.destination)
	c.Assert(err, IsNil)

	err = s.store.DeleteDestination(s.service, s.destination)
	c.Assert(err, IsNil)

	var expected []Destination
	destinations, err := s.store.GetDestinations(s.service)
	c.Assert(*destinations, DeepEquals, expected)
}
