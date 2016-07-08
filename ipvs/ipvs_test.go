package ipvs_test

import (
	"fmt"
	"sort"

	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/ipvs"

	. "gopkg.in/check.v1"
)

func (s *IpvsSuite) TestNewIpvs(c *C) {
	i, err := ipvs.New()
	c.Assert(err, IsNil)
	err = i.SyncState(s.state)
	c.Assert(err, IsNil)
}

type serviceList []*gipvs.Service

func (l serviceList) Len() int      { return len(l) }
func (l serviceList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l serviceList) Less(i, j int) bool {
	return fmt.Sprintf("%v", l[i].Address) < fmt.Sprintf("%v", l[j].Address)
}

func matchState(c *C, services1 []*gipvs.Service, state ipvs.State) {
	for _, s := range services1 {
		s.Flags = 0
		s.Statistics = nil
		for _, d := range s.Destinations {
			d.Statistics = nil
		}
	}
	stateServices := state.GetServices()
	cmp := make([]*gipvs.Service, len(stateServices))
	for i, s := range stateServices {
		cmp[i] = ipvs.ToIpvsService(&s)
	}
	sort.Sort(serviceList(services1))
	sort.Sort(serviceList(cmp))
	c.Check(services1, DeepEquals, cmp)
}

func (s *IpvsSuite) TestIpvsSyncState(c *C) {
	i, err := ipvs.New()
	c.Assert(err, IsNil)
	srv2 := &types.Service{
		Name:         "test1",
		Host:         "10.0.9.9",
		Port:         80,
		Scheduler:    "lc",
		Protocol:     "tcp",
		Destinations: []types.Destination{},
	}
	dst2 := &types.Destination{
		Name:   "test1",
		Host:   "192.168.9.9",
		Port:   80,
		Mode:   "nat",
		Weight: 1,
	}
	s.state.AddService(s.service)
	s.state.AddDestination(s.destination)
	err = i.SyncState(s.state)
	c.Assert(err, IsNil)
	services, err := gipvs.GetServices()
	c.Assert(err, IsNil)
	matchState(c, services, s.state)
	err = i.SyncState(s.state)
	c.Assert(err, IsNil)
	services, err = gipvs.GetServices()
	c.Assert(err, IsNil)
	matchState(c, services, s.state)
	s.state.DeleteDestination(s.destination)
	dst2.ServiceId = s.destination.ServiceId
	s.state.AddDestination(dst2)
	s.state.AddService(srv2)
	dst2.Name = "testx"
	dst2.ServiceId = srv2.GetId()
	s.state.AddDestination(dst2)
	err = i.SyncState(s.state)
	c.Assert(err, IsNil)
	services, err = gipvs.GetServices()
	c.Assert(err, IsNil)
	matchState(c, services, s.state)
	s.state.DeleteService(srv2)
	err = i.SyncState(s.state)
	c.Assert(err, IsNil)
	services, err = gipvs.GetServices()
	c.Assert(err, IsNil)
	matchState(c, services, s.state)
}
