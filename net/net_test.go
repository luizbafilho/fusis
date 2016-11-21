package net_test

import (
	"testing"

	"github.com/luizbafilho/fusis/net"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type NetSuite struct {
	iface string
}

var _ = Suite(&NetSuite{"lo"})

func (s *NetSuite) SetUpSuite(c *C) {
}

func (s *NetSuite) SetUpTest(c *C) {
	net.DelVips(s.iface)
}

func (s *NetSuite) TearDownTest(c *C) {
	net.DelVips("")
}

func (s *NetSuite) TestAddIp(c *C) {
	err := net.AddIp("192.168.0.1/32", "lo")
	c.Assert(err, IsNil)

	addrs, err := net.GetVips(s.iface)
	c.Assert(err, IsNil)

	found := false
	for _, a := range addrs {
		if a.IPNet.String() == "192.168.0.1/32" {
			found = true
		}
	}

	c.Assert(found, Equals, true)
}

func (s *NetSuite) TestDelIp(c *C) {
	err := net.AddIp("192.168.0.1/32", "lo")
	c.Assert(err, IsNil)

	err = net.DelIp("192.168.0.1/32", "lo")
	c.Assert(err, IsNil)

	addrs, err := net.GetVips(s.iface)
	c.Assert(err, IsNil)

	deleted := true
	for _, a := range addrs {
		if a.IPNet.String() == "192.168.0.1/32" {
			deleted = false
		}
	}

	c.Assert(deleted, Equals, true)
}

func (s *NetSuite) TestDelVips(c *C) {
	err := net.AddIp("192.168.0.1/32", "lo")
	c.Assert(err, IsNil)
	err = net.AddIp("192.168.0.2/32", "lo")
	c.Assert(err, IsNil)

	err = net.DelVips(s.iface)
	c.Assert(err, IsNil)

	addrs, err := net.GetVips(s.iface)
	c.Assert(err, IsNil)

	found := false
	for _, a := range addrs {
		if a.IPNet.String() == "192.168.0.1/32" || a.IPNet.String() == "192.168.0.2/32" {
			found = true
		}
	}

	c.Assert(found, Equals, false)
}

func (s *NetSuite) TestGetVips(c *C) {
	err := net.AddIp("192.168.0.1/32", "lo")
	c.Assert(err, IsNil)
	err = net.AddIp("192.168.0.2/32", "lo")
	c.Assert(err, IsNil)

	addrs, err := net.GetVips(s.iface)
	c.Assert(err, IsNil)

	c.Assert(len(addrs), Equals, 3)
}
