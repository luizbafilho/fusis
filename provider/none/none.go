package none

import (
	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider"
	"github.com/spf13/viper"
)

type None struct {
	Interface string
}

func init() {
	provider.RegisterProviderFactory("none", new)
}

func new() provider.Provider {
	viper.SetDefault("provider.params.interface", "eth0")

	return &None{
		Interface: viper.GetString("provider.params.interface"),
	}
}

func (n None) AllocateVip(s *ipvs.Service) error {
	// ip, err := n.getIp()
	// if err != nil {
	// 	return nil, err
	// }
	//
	// err = fusis_net.AddIp(ip, n.Interface)
	// if err != nil {
	// 	return nil, err
	// }
	return nil
}

func (n None) ReleaseVip(s ipvs.Service) error {
	logrus.Error("Deleting vip")
	return nil
}
