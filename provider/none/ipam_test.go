package none_test

import (
	"testing"

	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider/none"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpamSuite struct {
	state *ipvs.FusisState
	ipam  *none.Ipam
}

var _ = Suite(&IpamSuite{})

func (s *IpamSuite) SetUpSuite(c *C) {
	state := ipvs.NewFusisState()

	ipam, err := none.NewIpam("192.168.0.0/28", state)
	c.Assert(err, IsNil)

	s.ipam = ipam
	s.state = state
}

func (s *IpamSuite) TearDownSuite(c *C) {
}

func (s *IpamSuite) TestIpAllocation(c *C) {
	service := &ipvs.Service{
		Name: "test",
		Host: "192.168.0.2",
	}
	s.state.AddService(service)

	ip, err := s.ipam.Allocate()
	c.Assert(err, IsNil)
	c.Assert(ip, Equals, "192.168.0.1")

	service = &ipvs.Service{
		Name: "test2",
		Host: "192.168.0.1",
	}
	s.state.AddService(service)

	ip, err = s.ipam.Allocate()
	c.Assert(err, IsNil)
	c.Assert(ip, Equals, "192.168.0.3")
}
