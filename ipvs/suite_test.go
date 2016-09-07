package ipvs

import (
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/state"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpvsSuite struct {
	state       state.State
	service     *types.Service
	destination *types.Destination
}

var _ = Suite(&IpvsSuite{})

var nextPort = 15000

func getPort() int {
	p := nextPort
	nextPort++
	return p
}

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

	config := defaultConfig()
	state, err := state.New(&config)
	c.Assert(err, IsNil)

	s.state = *state
}

func (s *IpvsSuite) SetUpTest(c *C) {
	// s.state = ipvs.NewFusisState()
}

func (s *IpvsSuite) TearDownSuite(c *C) {
	// i, err := New()
	// c.Assert(err, IsNil)
	// err = i.Flush()
	// c.Assert(err, IsNil)
}

func defaultConfig() config.BalancerConfig {
	dir := tmpDir()
	return config.BalancerConfig{
		PublicInterface: "eth0",
		Name:            "Test",
		DataPath:        dir,
		Bootstrap:       true,
		Ports: map[string]int{
			"raft": getPort(),
			"serf": getPort(),
		},
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/28"},
		},
	}
}

func tmpDir() string {
	dir, _ := ioutil.TempDir("", "fusis")
	return dir
}
