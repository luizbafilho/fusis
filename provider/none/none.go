package none

import (
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipam"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/provider"
)

type None struct {
	Interface string
	VipRange  string
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

func (n None) Initialize() error {
	return ipam.Init(n.VipRange)
}

func (n None) AllocateVip(s *ipvs.Service) error {
	ip, err := ipam.Allocate()
	if err != nil {
		return err
	}
	s.Host = ip

	return net.AddIp(ip+"/32", config.Balancer.Provider.Params["interface"])
}

func (n None) ReleaseVip(s ipvs.Service) error {
	err := ipam.Release(s.Host)
	if err != nil {
		return err
	}

	return net.DelIp(s.Host+"/32", config.Balancer.Provider.Params["interface"])
}
