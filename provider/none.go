package provider

import (
	"fmt"

	"github.com/deckarep/golang-set"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state"
)

type None struct {
	iface string
	ipam  *Ipam
}

func NewNone(config *config.BalancerConfig) (Provider, error) {
	i, err := NewIpam(config.Provider.Params["vip-range"])
	if err != nil {
		return nil, err
	}

	return &None{
		iface: config.Provider.Params["interface"],
		ipam:  i,
	}, nil
}

func (n None) AllocateVIP(s *types.Service, state state.Store) error {
	ip, err := n.ipam.Allocate(state)
	if err != nil {
		return err
	}
	s.Host = ip

	return nil
}

func (n None) ReleaseVIP(s types.Service) error {
	n.ipam.Release(s.Host)
	return nil
}

func (n None) Sync(state state.State) error {
	currentVips, err := n.getCurrentVips()
	if err != nil {
		return err
	}

	stateVips := n.getStateVips(state)

	vipsToAdd := stateVips.Difference(currentVips)
	vipsToRemove := currentVips.Difference(stateVips)

	for v := range vipsToAdd.Iter() {
		vip := v.(string)
		err := net.AddIp(vip+"/32", n.iface)
		if err != nil {
			return fmt.Errorf("error adding ip %s: %s", vip, err)
		}
	}

	for v := range vipsToRemove.Iter() {
		vip := v.(string)
		err := net.DelIp(vip+"/32", n.iface)
		if err != nil {
			return fmt.Errorf("error deleting ip %s: %s", vip, err)
		}
	}

	return nil
}

func (n None) getCurrentVips() (mapset.Set, error) {
	vips, err := net.GetFusisVipsIps(n.iface)
	if err != nil {
		return nil, err
	}

	set := mapset.NewSet()
	for _, v := range vips {
		set.Add(v)
	}

	return set, nil
}

func (n None) getStateVips(state state.State) mapset.Set {
	vips := []string{}

	svcs := state.GetServices()
	for _, s := range svcs {
		vips = append(vips, s.Host)
	}

	set := mapset.NewSet()
	for _, v := range vips {
		set.Add(v)
	}

	return set
}
