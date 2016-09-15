package state_test

import (
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/state"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type StateSuite struct {
	state       state.Store
	service     *types.Service
	destination *types.Destination
}

var _ = Suite(&StateSuite{})

func (s *StateSuite) SetUpSuite(c *C) {
	logrus.SetOutput(ioutil.Discard)
	s.service = &types.Service{
		Name:      "test",
		Host:      "10.0.1.1",
		Port:      80,
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	s.destination = &types.Destination{
		Name:      "test",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *StateSuite) SetUpTest(c *C) {
	s.state = state.NewFusisStore()
}

func (s *StateSuite) TearDownSuite(c *C) {
	i, err := ipvs.New()
	c.Assert(err, IsNil)
	err = i.Flush()
	c.Assert(err, IsNil)
}
