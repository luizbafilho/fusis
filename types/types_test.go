package types

import (
	"testing"

	"gopkg.in/check.v1"
)

type S struct{}

var _ = check.Suite(&S{})

func Test(t *testing.T) { check.TestingT(t) }

func (s *S) TestServiceGetId(c *check.C) {
	srv := Service{Name: "myname"}
	c.Assert(srv.GetId(), check.Equals, "myname")
}

func (s *S) TestDestinationGetId(c *check.C) {
	dst := Destination{Name: "myname"}
	c.Assert(dst.GetId(), check.Equals, "myname")
}

func (s *S) TestErrors(c *check.C) {
	c.Assert(ErrServiceNotFound, check.FitsTypeOf, ErrNotFound(""))
	c.Assert(ErrDestinationNotFound, check.FitsTypeOf, ErrNotFound(""))
	c.Assert(ErrServiceNotFound.Error(), check.Equals, "service not found")
	c.Assert(ErrDestinationNotFound.Error(), check.Equals, "destination not found")
}
