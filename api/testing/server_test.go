package testing

import (
	"testing"

	"github.com/luizbafilho/fusis/api/types"
	"gopkg.in/check.v1"
)

type S struct {
}

var _ = check.Suite(&S{})

func Test(t *testing.T) { check.TestingT(t) }

func (s *S) TestNewFakeFusisServer(c *check.C) {
	srv := NewFakeFusisServer()
	c.Assert(srv.Balancer, check.FitsTypeOf, newTestBalancer())
	c.Assert(srv.api, check.NotNil)
}

func (s *S) TestTestBalancerAddService(c *check.C) {
	bal := newTestBalancer()
	srv := &types.Service{
		Name: "srv1",
	}
	err := bal.AddService(srv)
	c.Assert(err, check.IsNil)
	err = bal.AddService(srv)
	c.Assert(err, check.Equals, types.ErrServiceAlreadyExists)
}

func (s *S) TestTestBalancerGetService(c *check.C) {
	bal := newTestBalancer()
	srv := &types.Service{
		Name: "srv1",
	}
	err := bal.AddService(srv)
	c.Assert(err, check.IsNil)
	srv1, err := bal.GetService("srv1")
	c.Assert(err, check.IsNil)
	c.Assert(srv1, check.DeepEquals, srv)
	srv2, err := bal.GetService("srv2")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
	c.Assert(srv2, check.IsNil)
}

func (s *S) TestTestBalancerGetServices(c *check.C) {
	bal := newTestBalancer()
	srv := &types.Service{
		Name: "srv1",
	}
	err := bal.AddService(srv)
	c.Assert(err, check.IsNil)
	srv.Name = "srv2"
	err = bal.AddService(srv)
	c.Assert(err, check.IsNil)
	services := bal.GetServices()
	c.Assert(services, check.DeepEquals, []types.Service{{Name: "srv1"}, {Name: "srv2"}})
}

func (s *S) TestTestBalancerDeleteService(c *check.C) {
	bal := newTestBalancer()
	srv := &types.Service{
		Name: "srv1",
	}
	err := bal.AddService(srv)
	c.Assert(err, check.IsNil)
	err = bal.DeleteService("srv1")
	c.Assert(err, check.IsNil)
	srv1, err := bal.GetService("srv1")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
	c.Assert(srv1, check.IsNil)
	err = bal.DeleteService("srv1")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
}

func (s *S) TestTestBalancerAddDestination(c *check.C) {
	bal := newTestBalancer()
	srv := &types.Service{
		Name: "srv1",
	}
	err := bal.AddService(srv)
	c.Assert(err, check.IsNil)

	dst := &types.Destination{
		Name:      "dst1",
		ServiceId: "srv1",
	}
	err = bal.AddDestination(srv, dst)
	c.Assert(err, check.IsNil)
	err = bal.AddDestination(srv, dst)
	c.Assert(err, check.Equals, types.ErrDestinationAlreadyExists)

	dst.Name = "dstX"
	dst.ServiceId = "srvX"
	srv.Name = "srvX"
	err = bal.AddDestination(srv, dst)
	c.Assert(err, check.Equals, types.ErrServiceNotFound)

	srv, err = bal.GetService("srv1")
	c.Assert(err, check.IsNil)
	c.Assert(srv.Destinations, check.DeepEquals, []types.Destination{{
		Name:      "dst1",
		ServiceId: "srv1",
	}})
}

func (s *S) TestTestBalancerGetDestination(c *check.C) {
	bal := newTestBalancer()
	srv := &types.Service{
		Name: "srv1",
	}
	err := bal.AddService(srv)
	c.Assert(err, check.IsNil)
	dst := &types.Destination{
		Name: "dst1",
	}
	err = bal.AddDestination(srv, dst)
	c.Assert(err, check.IsNil)
	dst1, err := bal.GetDestination("dst1")
	c.Assert(err, check.IsNil)
	c.Assert(dst1, check.DeepEquals, dst)
	dst2, err := bal.GetDestination("dst2")
	c.Assert(err, check.Equals, types.ErrDestinationNotFound)
	c.Assert(dst2, check.IsNil)
}

func (s *S) TestTestBalancerDeleteDestination(c *check.C) {
	bal := newTestBalancer()
	srv := &types.Service{
		Name: "srv1",
	}
	err := bal.AddService(srv)
	c.Assert(err, check.IsNil)

	dst := &types.Destination{
		Name:      "dst1",
		ServiceId: "srv1",
		Host:      "192.1.1.1",
	}
	err = bal.AddDestination(srv, dst)
	c.Assert(err, check.IsNil)

	dst2 := &types.Destination{
		Name:      "dst2",
		ServiceId: "srv1",
		Host:      "192.2.2.2",
	}
	err = bal.AddDestination(srv, dst2)
	c.Assert(err, check.IsNil)

	err = bal.DeleteDestination(dst)
	c.Assert(err, check.IsNil)
	err = bal.DeleteDestination(dst)
	c.Assert(err, check.Equals, types.ErrDestinationNotFound)

	dst, err = bal.GetDestination("dst2")
	c.Assert(err, check.IsNil)
	c.Assert(dst, check.DeepEquals, &types.Destination{
		Name:      "dst2",
		ServiceId: "srv1",
		Host:      "192.2.2.2",
	})

	srv, err = bal.GetService("srv1")
	c.Assert(err, check.IsNil)
	c.Assert(srv.Destinations, check.DeepEquals, []types.Destination{{
		Name:      "dst2",
		ServiceId: "srv1",
		Host:      "192.2.2.2",
	}})

	dst.Name = "dst1"
	err = bal.DeleteDestination(dst)
	c.Assert(err, check.IsNil)

	srv, err = bal.GetService("srv1")
	c.Assert(err, check.IsNil)
	c.Assert(srv.Destinations, check.DeepEquals, []types.Destination{})
}
