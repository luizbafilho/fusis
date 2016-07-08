package fusis

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/engine"
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
	service     *types.Service
	destination *types.Destination
	config      *config.BalancerConfig
}

var _ = Suite(&FusisSuite{})

func (s *FusisSuite) SetUpSuite(c *C) {
	// logrus.SetOutput(ioutil.Discard)
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
		Name:       "Test",
		ConfigPath: dir,
		Bootstrap:  true,
		Ports: map[string]int{
			"raft": getPort(),
			"serf": getPort(),
		},
		Provider: config.Provider{
			Type: "none",
			Params: map[string]string{
				"interface": "eth0",
				"vipRange":  "192.168.0.0/28",
			},
		},
	}
}

func (s *FusisSuite) TestAddService(c *C) {
	config := defaultConfig()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)
	WaitForResult(func() (bool, error) {
		return b.isLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})
	srv, err := b.GetService(s.service.Name)
	c.Assert(err, Equals, types.ErrServiceNotFound)
	err = b.AddService(s.service)
	c.Assert(err, IsNil)
	srv, err = b.GetService(s.service.Name)
	c.Assert(err, IsNil)
	c.Assert(srv, DeepEquals, s.service)
	err = b.AddService(s.service)
	c.Assert(err, Equals, types.ErrServiceAlreadyExists)
}

func (s *FusisSuite) TestAddServiceConcurrent(c *C) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(10))
	config := defaultConfig()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)
	WaitForResult(func() (bool, error) {
		return b.isLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})
	n := 100
	errs := make([]error, n)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = b.AddService(s.service)
		}(i)
	}
	wg.Wait()
	count := 0
	for _, err := range errs {
		if err == nil {
			count++
		} else {
			c.Assert(err, Equals, types.ErrServiceAlreadyExists)
		}
	}
	c.Assert(count, Equals, 1)
	srv, err := b.GetService(s.service.Name)
	c.Assert(err, IsNil)
	c.Assert(srv, DeepEquals, s.service)
}

func (s *FusisSuite) TestDeleteService(c *C) {
	config := defaultConfig()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)
	WaitForResult(func() (bool, error) {
		return b.isLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})
	err = b.AddService(s.service)
	c.Assert(err, IsNil)
	err = b.DeleteService(s.service.Name)
	c.Assert(err, IsNil)
	err = b.DeleteService(s.service.Name)
	c.Assert(err, Equals, types.ErrServiceNotFound)
}

func (s *FusisSuite) TestDeleteServiceConcurrent(c *C) {
	config := defaultConfig()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)
	WaitForResult(func() (bool, error) {
		return b.isLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})
	err = b.AddService(s.service)
	c.Assert(err, IsNil)
	n := 100
	errs := make([]error, n)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("test-%d", i)
			bErr := b.AddService(&types.Service{
				Name:         name,
				Port:         80,
				Scheduler:    "lc",
				Protocol:     "tcp",
				Destinations: []types.Destination{},
			})
			c.Assert(bErr, IsNil)
			bErr = b.DeleteService(name)
			c.Assert(bErr, IsNil)
			errs[i] = b.DeleteService(s.service.Name)
		}(i)
	}
	wg.Wait()
	count := 0
	for _, err := range errs {
		if err == nil {
			count++
		} else {
			c.Assert(err, Equals, types.ErrServiceNotFound)
		}
	}
	c.Assert(count, Equals, 1)
	all := b.GetServices()
	c.Assert(all, DeepEquals, []types.Service{})
}

func (s *FusisSuite) TestAssignVIP(c *C) {
	config := defaultConfig()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)

	WaitForResult(func() (bool, error) {
		return b.IsLeader(), nil
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
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)

	WaitForResult(func() (bool, error) {
		return b.IsLeader(), nil
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

func (s *FusisSuite) TestJoinPoolLeave(c *C) {
	config := defaultConfig()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)

	WaitForResult(func() (bool, error) {
		return b.IsLeader(), nil
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

	// Testing JoinPool
	err = s2.JoinPool()
	c.Assert(err, IsNil)

	// Check the members
	WaitForResult(func() (bool, error) {
		return len(s2.serf.Members()) == 2, nil
	}, func(err error) {
		c.Fatalf("balancer could not join the serf cluster")
	})

	// Check the members
	WaitForResult(func() (bool, error) {
		peers, _ := b.raftPeers.Peers()
		return len(peers) == 2, nil
	}, func(err error) {
		c.Fatalf("balancer could not join the raft cluster")
	})

	// Testing Leave Pool
	s2.Leave()

	WaitForResult(func() (bool, error) {
		peers, _ := b.raftPeers.Peers()
		return len(peers) == 1, nil
	}, func(err error) {
		c.Fatalf("balancer could not leave the raft cluster")
	})
}

func (s *FusisSuite) TestWatchCommands(c *C) {
	config := defaultConfig()
	b, err := NewBalancer(&config)
	c.Assert(err, IsNil)
	defer b.Shutdown()
	defer os.RemoveAll(config.ConfigPath)

	WaitForResult(func() (bool, error) {
		return b.IsLeader(), nil
	}, func(err error) {
		c.Fatalf("balancer did not become leader")
	})

	s.service.Host = "192.168.85.43"

	// Sending add service event to balancer assign the VIP
	cmd := engine.Command{
		Op:      engine.AddServiceOp,
		Service: s.service,
	}
	b.engine.CommandCh <- cmd

	WaitForResult(func() (bool, error) {
		addrs, err := net.GetVips(config.Interface)
		c.Assert(err, IsNil)

		found := false
		for _, a := range addrs {
			if a.IPNet.String() == "192.168.85.43/32" {
				found = true
			}
		}

		return found, nil
	}, func(err error) {
		c.Fatalf("balancer did not assigned ip")
	})

	// Sending del service event to balancer unassign the VIP
	cmd = engine.Command{
		Op:      engine.DelServiceOp,
		Service: s.service,
	}
	b.engine.CommandCh <- cmd

	WaitForResult(func() (bool, error) {
		addrs, err := net.GetVips(config.Interface)
		c.Assert(err, IsNil)

		deleted := true
		for _, a := range addrs {
			if a.IPNet.String() == "192.168.85.43/32" {
				deleted = false
			}
		}

		return deleted, nil
	}, func(err error) {
		c.Fatalf("balancer did not unassigned ip")
	})
}
