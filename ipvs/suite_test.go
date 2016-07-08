package ipvs_test

import (
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/ipvs"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpvsSuite struct {
	state       ipvs.State
	service     *types.Service
	destination *types.Destination
}

var _ = Suite(&IpvsSuite{})

func (s *IpvsSuite) SetUpSuite(c *C) {
	logrus.SetOutput(ioutil.Discard)
	s.service = &types.Service{
		Name:         "test",
		Host:         "10.0.1.1",
		Port:         80,
		Scheduler:    "lc",
		Protocol:     "tcp",
		Destinations: []types.Destination{},
	}

	s.destination = &types.Destination{
		Name:      "test",
		Host:      "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *IpvsSuite) SetUpTest(c *C) {
	s.state = ipvs.NewFusisState()
}

func (s *IpvsSuite) TearDownSuite(c *C) {
	i, err := ipvs.New()
	c.Assert(err, IsNil)
	err = i.Flush()
	c.Assert(err, IsNil)
}
