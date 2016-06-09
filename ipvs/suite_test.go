package ipvs_test

import (
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/ipvs"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpvsSuite struct {
	state       ipvs.State
	service     *ipvs.Service
	destination *ipvs.Destination
}

var _ = Suite(&IpvsSuite{})

func (s *IpvsSuite) SetUpSuite(c *C) {
	logrus.SetOutput(ioutil.Discard)

	s.service = &ipvs.Service{
		Name:         "test",
		Host:         "10.0.1.1",
		Port:         80,
		Scheduler:    "lc",
		Protocol:     "tcp",
		Destinations: []ipvs.Destination{},
	}

	s.destination = &ipvs.Destination{
		Name:      "test",
		Host:      "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *IpvsSuite) SetUpTest(c *C) {
	state := ipvs.NewFusisState()

	s.state = state
}
