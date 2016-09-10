package state_test

import (
	"github.com/luizbafilho/fusis/api/types"
	. "gopkg.in/check.v1"
)

func (s *StateSuite) TestGetService(c *C) {
	s.state.AddService(s.service)
	s.state.AddDestination(s.destination)

	svcs := s.state.GetServices()
	// s.service.Destinations = []types.Destination{*s.destination}
	c.Assert(svcs[0], DeepEquals, *s.service)

	svc, err := s.state.GetService(s.service.Name)
	c.Assert(err, IsNil)
	c.Assert(svc, DeepEquals, s.service)

	_, err = s.state.GetService("unknown")
	c.Assert(err, Equals, types.ErrServiceNotFound)
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

	services := s.state.GetServices()
	c.Assert(len(services), Equals, 0)

	_, err := s.state.GetService(s.service.Name)
	c.Assert(err, Equals, types.ErrServiceNotFound)
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
	c.Assert(err, DeepEquals, types.ErrDestinationNotFound)
}
