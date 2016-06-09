package ipvs_test

import (
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/ipvs"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type StateSuite struct {
	state       ipvs.State
	service     *ipvs.Service
	destination *ipvs.Destination
}

var _ = Suite(&StateSuite{})

func (s *StateSuite) SetUpSuite(c *C) {
	logrus.SetOutput(ioutil.Discard)

	s.service = &ipvs.Service{
		Name:         "test",
		Host:         "10.0.1.1",
		Port:         80,
		Scheduler:    "lc",
		Protocol:     "tcp",
		Destinations: []ipvs.Destination{},
	}

	s.destination = &ipvs.Destination{
		Name:      "test",
		Host:      "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *StateSuite) SetUpTest(c *C) {
	state := ipvs.NewFusisState()

	s.state = state
}

func (s *StateSuite) TestGetService(c *C) {
	s.state.AddService(s.service)
	s.state.AddDestination(s.destination)

	svcs := *s.state.GetServices()
	s.service.Destinations = []ipvs.Destination{*s.destination}
	c.Assert(svcs[0], DeepEquals, *s.service)

	svc, err := s.state.GetService(s.service.Name)
	c.Assert(err, IsNil)
	c.Assert(svc, DeepEquals, s.service)

	_, err = s.state.GetService("unknown")
	c.Assert(err, Equals, ipvs.ErrNotFound)
}

func (s *StateSuite) TestAddService(c *C) {
	s.state.AddService(s.service)

	service, err := s.state.GetService(s.service.Name)
	c.Assert(err, IsNil)
	c.Assert(service, DeepEquals, s.service)
}

func (s *StateSuite) TestDelService(c *C) {
	s.state.AddService(s.service)
	s.state.DeleteService(s.service)

	services := *s.state.GetServices()
	c.Assert(len(services), Equals, 0)

	_, err := s.state.GetService(s.service.Name)
	c.Assert(err, Equals, ipvs.ErrNotFound)
}

func (s *StateSuite) TestAddDestination(c *C) {
	s.state.AddService(s.service)
	s.state.AddDestination(s.destination)

	dst, err := s.state.GetDestination(s.destination.Name)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, s.destination)
}

func (s *StateSuite) TestDelDestination(c *C) {
	s.state.AddService(s.service)
	s.state.AddDestination(s.destination)
	s.state.DeleteDestination(s.destination)

	_, err := s.state.GetDestination(s.destination.Name)
	c.Assert(err, DeepEquals, ipvs.ErrNotFound)
}
