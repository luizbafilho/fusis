package provider

import (
	"fmt"
	"strings"

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

//Sync syncs the given state by setting all vips on network interface
func (n None) Sync(state state.State) error {
	oldVIPs, err := net.GetFusisVipsIps(n.iface)
	if err != nil {
		return err
	}

	newServices := state.GetServices()
	toAddMap := make(map[string]struct{})
	for _, s := range newServices {
		toAddMap[s.Host] = struct{}{}
	}

	var toRemove []string
	for _, ip := range oldVIPs {
		if _, isPresent := toAddMap[ip]; isPresent {
			delete(toAddMap, ip)
		} else {
			toRemove = append(toRemove, ip)
		}
	}

	var errors []string
	for ip := range toAddMap {
		err := net.AddIp(ip+"/32", n.iface)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error adding ip %s: %s", ip, err))
		}
	}

	for _, ip := range toRemove {
		err := net.DelIp(ip+"/32", n.iface)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error deleting ip %s: %s", ip, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors: %s", strings.Join(errors, " | "))
	}

	return nil
}
