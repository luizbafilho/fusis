package steps

import (
	"testing"

	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestSuccess(c *check.C) {
	s1 := step1{}
	s2 := step2{}

	seq := NewSequence(s1, s2)
	err := seq.Execute()

	c.Assert(err, check.IsNil)
}

func (s *S) TestError(c *check.C) {
	s1 := step1{}
	s2 := errorStep{}

	seq := NewSequence(s1, s2)
	err := seq.Execute()

	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "Error")
}
