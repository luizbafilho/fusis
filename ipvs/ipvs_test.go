package ipvs_test

import (
	"github.com/luizbafilho/fusis/ipvs"

	. "gopkg.in/check.v1"
)

func (s *IpvsSuite) TestNewIpvs(c *C) {
	i := ipvs.New()

	err := i.AddService(s.service.ToIpvsService())
	c.Assert(err, IsNil)

	// Verifies if a new instance flushes the ipvs table
	i = ipvs.New()
	err = i.AddService(s.service.ToIpvsService())
	c.Assert(err, IsNil)
}
