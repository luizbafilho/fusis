package ipam_test

import (
	"testing"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipam"
	"github.com/luizbafilho/fusis/state"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpamSuite struct {
	state *state.State
	ipam  ipam.Allocator
}

var _ = Suite(&IpamSuite{})

func (s *IpamSuite) SetUpSuite(c *C) {
	var err error
	config := &config.BalancerConfig{
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/28"},
		},
	}

	s.state, err = state.New(config)
	c.Assert(err, IsNil)

	s.ipam, err = ipam.New(s.state, config)
	c.Assert(err, IsNil)
}

func (s *IpamSuite) TearDownSuite(c *C) {
}

func (s *IpamSuite) TestIpAllocation(c *C) {
	service := &types.Service{
		Name: "test",
		Host: "192.168.0.2",
	}
	s.state.AddService(service)

	err := s.ipam.AllocateVIP(service)
	c.Assert(err, IsNil)
	c.Assert(service.Host, Equals, "192.168.0.1")

	service = &types.Service{
		Name: "test2",
		Host: "192.168.0.1",
	}
	s.state.AddService(service)

	err = s.ipam.AllocateVIP(service)
	c.Assert(err, IsNil)
	c.Assert(service.Host, Equals, "192.168.0.3")
}
