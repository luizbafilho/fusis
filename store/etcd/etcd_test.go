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
	store   Store
	service Service
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
	svc, err := s.store.GetService(s.service.GetId())

	c.Assert(err, IsNil)
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
