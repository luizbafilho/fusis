package fusis

import (
	"fmt"
	"testing"
	"time"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
	. "gopkg.in/check.v1"
)

var nextPort = 15000

func getPort() int {
	p := nextPort
	nextPort++
	return p
}

func Test(t *testing.T) { TestingT(t) }

type FusisSuite struct {
	balancer    *Balancer
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
		DevMode: true,
		Single:  true,
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

type testFn func() (bool, error)
type errorFn func(error)

func WaitForResult(test testFn, error errorFn) {
	retries := 1000

	for retries > 0 {
		time.Sleep(10 * time.Millisecond)
		retries--

		success, err := test()
		if success {
			return
		}

		if retries == 0 {
			error(err)
		}
	}
}

func (s *FusisSuite) TestGetServices(c *C) {
}

func (s *FusisSuite) TestAddService(c *C) {
	config := *s.config
	config.RaftPort = getPort()
	config.SerfPort = getPort()
	_, err := NewBalancer(&config)
	c.Assert(err, IsNil)

	join, err := s.config.GetIpByInterface()
	c.Assert(err, IsNil)

	config2 := *s.config
	config2.Name = "test"
	config2.RaftPort = getPort()
	config2.SerfPort = getPort()
	config2.Join = fmt.Sprintf("%v:%v", join, config.SerfPort)
	config2.Single = false

	s2, err := NewBalancer(&config2)
	c.Assert(err, IsNil)

	err = s2.JoinPool()
	c.Assert(err, IsNil)

	// Check the members
	WaitForResult(func() (bool, error) {
		return len(s2.serf.Members()) == 2, nil
	}, func(err error) {
		c.Fatalf("bad len")
	})

	// c.Assert(len(s2.serf.Members()), Equals, 2)
}
