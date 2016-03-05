package etcd_test

import (
	"testing"

	"github.com/luizbafilho/janus/store"
	"github.com/luizbafilho/janus/store/etcd"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type EtcdSuite struct {
	store store.Store
}

var _ = Suite(&EtcdSuite{})

func (s *EtcdSuite) SetUpTest(c *C) {
	nodes := []string{"http://127.0.0.1:2379"}
	s.store = etcd.New(nodes)
}

func (s *EtcdSuite) TestGetServices(c *C) {
	svcs, _ := s.store.GetServices()

	expectedService := store.Service{
		Host:      "10.7.0.1",
		Port:      80,
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	expected := []store.Service{expectedService}

	c.Assert(*svcs, DeepEquals, expected)
}
