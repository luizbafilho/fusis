package api_test

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"sort"
	"time"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/fusis"
	"gopkg.in/check.v1"
)

func (s *S) TestFullstackWithClient(c *check.C) {
	if os.Getenv("TRAVIS") == "true" {
		c.Skip("Skipping test because travis-ci do not allow iptables")
	}

	dir, err := ioutil.TempDir("", "fusis")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(dir)
	conf := config.BalancerConfig{
		Interfaces: config.Interfaces{
			Inbound:  "eth0",
			Outbound: "eth0",
		},
		Name:      "Test",
		DataPath:  dir,
		Bootstrap: true,
		Ports: map[string]int{
			"raft": 20012,
			"serf": 20013,
		},
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/28"},
		},
	}
	balancer, err := fusis.NewBalancer(&conf)
	c.Assert(err, check.IsNil)
	defer balancer.Shutdown()
	timeout := time.After(30 * time.Second)
	for {
		if balancer.IsLeader() {
			break
		}
		select {
		case <-time.After(10 * time.Millisecond):
		case <-timeout:
			c.Fatal("timeout waiting for leader after 30 seconds")
		}
	}
	apiHandler := api.NewAPI(balancer)
	srv := httptest.NewServer(apiHandler)
	client := api.NewClient(srv.URL)
	_, err = client.CreateService(types.Service{Name: "myservice", Port: 1040, Mode: "nat", Protocol: "tcp", Scheduler: "rr"})
	c.Assert(err, check.IsNil)
	_, err = client.CreateService(types.Service{Name: "myservice", Port: 1050, Mode: "nat", Protocol: "tcp", Scheduler: "rr"})
	c.Assert(err, check.Equals, types.ErrServiceAlreadyExists)
	_, err = client.AddDestination(types.Destination{ServiceId: "myservice", Name: "myname1", Address: "10.0.0.1", Port: 1234, Mode: "nat"})
	c.Assert(err, check.IsNil)
	_, err = client.AddDestination(types.Destination{ServiceId: "myservice", Name: "myname2", Address: "10.0.0.2", Port: 1234, Mode: "nat"})
	c.Assert(err, check.IsNil)
	_, err = client.AddDestination(types.Destination{ServiceId: "myservice", Name: "myname3", Address: "10.0.0.1", Port: 1235, Mode: "nat"})
	c.Assert(err, check.IsNil)
	_, err = client.AddDestination(types.Destination{ServiceId: "myservice", Name: "myname3", Address: "10.0.0.1", Port: 1235, Mode: "nat"})
	c.Assert(err, check.Equals, types.ErrDestinationAlreadyExists)
	_, err = client.AddDestination(types.Destination{ServiceId: "myservice", Name: "myname4", Address: "10.0.0.1", Port: 1234, Mode: "nat"})
	c.Assert(err, check.Equals, types.ErrDestinationAlreadyExists)
	_, err = client.AddDestination(types.Destination{ServiceId: "myserviceX", Name: "myname3", Address: "10.0.0.1", Port: 1235, Mode: "nat"})
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
	services, err := client.GetServices()
	c.Assert(err, check.IsNil)
	c.Assert(services, check.HasLen, 1)
	svc := *services[0]
	c.Assert(svc, check.DeepEquals, types.Service{
		Name:      "myservice",
		Mode:      "nat",
		Port:      1040,
		Protocol:  "tcp",
		Scheduler: "rr",
		Host:      "192.168.0.1",
	})
	dsts := balancer.GetDestinations(&svc)
	sort.Sort(types.DestinationList(dsts))
	c.Assert(dsts, check.DeepEquals, []types.Destination{
		{
			Name:      "myname1",
			Address:   "10.0.0.1",
			Port:      1234,
			Weight:    1,
			Mode:      "nat",
			ServiceId: "myservice",
		},
		{
			Name:      "myname2",
			Address:   "10.0.0.2",
			Port:      1234,
			Weight:    1,
			Mode:      "nat",
			ServiceId: "myservice",
		},
		{
			Name:      "myname3",
			Address:   "10.0.0.1",
			Port:      1235,
			Weight:    1,
			Mode:      "nat",
			ServiceId: "myservice",
		},
	})
	err = client.DeleteDestination("myservice", "myname2")
	c.Assert(err, check.IsNil)
	err = client.DeleteDestination("myservice", "myname3")
	c.Assert(err, check.IsNil)
	err = client.DeleteDestination("myservice", "myname3")
	c.Assert(err, check.Equals, types.ErrDestinationNotFound)
	err = client.DeleteDestination("myserviceX", "myname2")
	c.Assert(err, check.Equals, types.ErrDestinationNotFound)
	services, err = client.GetServices()
	c.Assert(err, check.IsNil)
	c.Assert(services, check.HasLen, 1)
	c.Assert(*services[0], check.DeepEquals, types.Service{
		Name:      "myservice",
		Mode:      "nat",
		Port:      1040,
		Protocol:  "tcp",
		Scheduler: "rr",
		Host:      "192.168.0.1",
	})
	c.Assert(balancer.GetDestinations(services[0]), check.DeepEquals, []types.Destination{
		{
			Name:      "myname1",
			Address:   "10.0.0.1",
			Port:      1234,
			Weight:    1,
			Mode:      "nat",
			ServiceId: "myservice",
		},
	})
	err = client.DeleteService("myserviceX")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
	err = client.DeleteService("myservice")
	c.Assert(err, check.IsNil)
	err = client.DeleteService("myservice")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
	services, err = client.GetServices()
	c.Assert(err, check.IsNil)
	c.Assert(services, check.HasLen, 0)
}
