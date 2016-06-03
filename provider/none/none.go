package none

import (
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/provider"
)

type None struct {
	Interface string
	VipRange  string
	ipam      *Ipam
}

func init() {
	provider.RegisterProviderFactory("none", new)
}

func new() provider.Provider {

	return &None{
		Interface: config.Balancer.Provider.Params["interface"],
		VipRange:  config.Balancer.Provider.Params["vipRange"],
	}
}

func (n *None) Initialize(state ipvs.State) error {
	i, err := NewIpam(n.VipRange, state)
	if err != nil {
		return err
	}

	n.ipam = i

	return nil
}

func (n None) AllocateVIP(s *ipvs.Service) error {
	ip, err := n.ipam.Allocate()
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

func (n None) AssignVIP(s ipvs.Service) error {
	return net.AddIp(s.Host+"/32", config.Balancer.Provider.Params["interface"])
}

func (n None) UnassignVIP(s ipvs.Service) error {
	return net.DelIp(s.Host+"/32", config.Balancer.Provider.Params["interface"])
}
