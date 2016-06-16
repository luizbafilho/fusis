package provider_test

import (
	"testing"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpamSuite struct {
	state *ipvs.FusisState
	ipam  *provider.Ipam
}

var _ = Suite(&IpamSuite{})

func (s *IpamSuite) SetUpSuite(c *C) {
	s.state = ipvs.NewFusisState()

	ipam, err := provider.NewIpam("192.168.0.0/28")
	c.Assert(err, IsNil)

	s.ipam = ipam
}

func (s *IpamSuite) TearDownSuite(c *C) {
}

func (s *IpamSuite) TestIpAllocation(c *C) {
	service := &types.Service{
		Name: "test",
		Host: "192.168.0.2",
	}
	s.state.AddService(service)

	ip, err := s.ipam.Allocate(s.state)
	c.Assert(err, IsNil)
	c.Assert(ip, Equals, "192.168.0.1")

	service = &types.Service{
		Name: "test2",
		Host: "192.168.0.1",
	}
	s.state.AddService(service)

	ip, err = s.ipam.Allocate(s.state)
	c.Assert(err, IsNil)
	c.Assert(ip, Equals, "192.168.0.3")
}
