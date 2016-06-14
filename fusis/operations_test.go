package fusis

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/net"
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
}

func (s *FusisSuite) SetUpTest(c *C) {
}

func tmpDir() string {
	dir, _ := ioutil.TempDir("", "fusis")
	return dir
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

func defaultConfig() config.BalancerConfig {
	dir := tmpDir()
	return config.BalancerConfig{
		Interface:  "eth0",
		ConfigPath: dir,
		Bootstrap:  true,
		Ports:      make(map[string]int),
		Provider: config.Provider{
			Type: "none",
			Params: map[string]string{
				"interface": "eth0",
				"vipRange":  "192.168.0.0/28",
			},
		},
	}
}

func (s *FusisSuite) TestGetServices(c *C) {
}

func (s *FusisSuite) TestAssignVIP(c *C) {
	config := defaultConfig()
	config.Name = "test"
	config.Ports["raft"] = getPort()
	config.Ports["serf"] = getPort()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)

	WaitForResult(func() (bool, error) {
		return b.isLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})

	s.service.Host = "192.168.85.43"

	b.AssignVIP(s.service)

	addrs, err := net.GetVips(config.Interface)
	c.Assert(err, IsNil)

	found := false
	for _, a := range addrs {
		if a.IPNet.String() == "192.168.85.43/32" {
			found = true
		}
	}

	c.Assert(found, Equals, true)
}

func (s *FusisSuite) TestUnassignVIP(c *C) {
	config := defaultConfig()
	config.Name = "test"
	config.Ports["raft"] = getPort()
	config.Ports["serf"] = getPort()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)

	WaitForResult(func() (bool, error) {
		return b.isLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})

	s.service.Host = "192.168.85.43"

	b.AssignVIP(s.service)
	b.UnassignVIP(s.service)

	addrs, err := net.GetVips(config.Interface)
	c.Assert(err, IsNil)

	deleted := true
	for _, a := range addrs {
		if a.IPNet.String() == "192.168.0.1/32" {
			deleted = false
		}
	}

	c.Assert(deleted, Equals, true)
}

func (s *FusisSuite) TestJoinPool(c *C) {
	config := defaultConfig()
	config.Name = "test"
	config.Ports["raft"] = getPort()
	config.Ports["serf"] = getPort()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)

	WaitForResult(func() (bool, error) {
		return b.isLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})

	join, err := config.GetIpByInterface()
	c.Assert(err, IsNil)

	config2 := defaultConfig()
	config2.Name = "test2"
	config2.Ports["raft"] = getPort()
	config2.Ports["serf"] = getPort()
	config2.Join = []string{fmt.Sprintf("%v:%v", join, config.Ports["serf"])}
	config2.Bootstrap = false

	s2, err := NewBalancer(&config2)
	c.Assert(err, IsNil)
	defer s2.Shutdown()
	defer os.RemoveAll(config2.ConfigPath)

	err = s2.JoinPool()
	c.Assert(err, IsNil)

	// Check the members
	WaitForResult(func() (bool, error) {
		return len(s2.serf.Members()) == 2, nil
	}, func(err error) {
		c.Fatalf("bad serf len")
	})

	// Check the members
	WaitForResult(func() (bool, error) {
		peers, _ := b.raftPeers.Peers()
		return len(peers) == 2, nil
	}, func(err error) {
		c.Fatalf("bad raft len")
	})
}
