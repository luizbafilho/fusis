package provider

import (
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/net"
)

type None struct {
	iface string
	ipam  *Ipam
}

func NewNone(config *config.BalancerConfig) (Provider, error) {
	i, err := NewIpam(config.Provider.Params["vipRange"])
	if err != nil {
		return nil, err
	}

	return &None{
		iface: config.Provider.Params["interface"],
		ipam:  i,
	}, nil
}

func (n None) AllocateVIP(s *ipvs.Service, state ipvs.State) error {
	ip, err := n.ipam.Allocate(state)
	if err != nil {
		return err
	}
	s.Host = ip

	return nil
}

func (n None) ReleaseVIP(s ipvs.Service) error {
	n.ipam.Release(s.Host)
	return nil
}

func (n None) AssignVIP(s *ipvs.Service) error {
	return net.AddIp(s.Host+"/32", n.iface)
}

func (n None) UnassignVIP(s *ipvs.Service) error {
	return net.DelIp(s.Host+"/32", n.iface)
}
