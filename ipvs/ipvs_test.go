package ipvs_test

import (
	"github.com/luizbafilho/fusis/ipvs"

	. "gopkg.in/check.v1"
)

func (s *IpvsSuite) TestNewIpvs(c *C) {
	i, err := ipvs.New()
	c.Assert(err, IsNil)

	err = i.AddService(ipvs.ToIpvsService(s.service))
	c.Assert(err, IsNil)

	// Verifies if a new instance flushes the ipvs table
	i, err = ipvs.New()
	c.Assert(err, IsNil)
	err = i.AddService(ipvs.ToIpvsService(s.service))
	c.Assert(err, IsNil)
}
