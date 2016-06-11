package fusis_test

import (
	"testing"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/fusis"
	"github.com/luizbafilho/fusis/ipvs"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FusisSuite struct {
	balancer    *fusis.Balancer
	service     *ipvs.Service
	destination *ipvs.Destination
	config      *config.BalancerConfig
}

var _ = Suite(&FusisSuite{})

func (s *FusisSuite) SetUpSuite(c *C) {
	// logrus.SetOutput(ioutil.Discard)
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

	config := &config.BalancerConfig{
		Single: true,
		Provider: config.Provider{
			Type: "none",
			Params: map[string]string{
				"interface": "eth0",
				"vipRange":  "192.168.0.0/28",
			},
		},
	}

	config.Interface = "eth0"
	s.config = config
}

func (s *FusisSuite) SetUpTest(c *C) {
}

func (s *FusisSuite) readConfig() {
}

func (s *FusisSuite) TestGetServices(c *C) {
}

func (s *FusisSuite) TestAddService(c *C) {
	// fusis.NewBalancer(s.config)
	// err := s.balancer.AddService(s.service)
	// c.Assert(err, IsNil)
	//
	// services := s.balancer.GetServices()
	// c.Assert(len(*services), DeepEquals, 1)
}
