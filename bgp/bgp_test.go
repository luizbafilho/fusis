package bgp

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type BgpSuite struct {
	service *BgpService
}

var _ = Suite(&BgpSuite{})

func (s *BgpSuite) SetUpSuite(c *C) {
}

func (s *BgpSuite) SetUpTest(c *C) {
}

func (s *BgpSuite) TearDownTest(c *C) {
}

func (s *BgpSuite) TestAddIp(c *C) {
}
