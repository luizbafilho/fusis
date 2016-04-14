package provider

import (
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"
	fusis_net "github.com/luizbafilho/fusis/net"
	"github.com/spf13/viper"
)

type None struct {
	Interface string
}

func init() {
	RegisterProviderFactory("none", newNoneProvider)
}

func newNoneProvider() Provider {
	viper.SetDefault("provider.params.interface", "eth0")

	return None{
		Interface: viper.GetString("provider.params.interface"),
	}
}

func (n None) SetVip() (interface{}, error) {
	iface, err := net.InterfaceByName(n.Interface)
	if err != nil {
		return nil, err
	}
	client, err := dhcp4client.New(dhcp4client.HardwareAddr(iface.HardwareAddr), dhcp4client.Timeout(30*time.Second))
	if err != nil {
		logrus.Errorf("can't create dhcp4 client")
		return nil, err
	}
	defer client.Close()
	ok, packet, err := client.Request()
	if !ok || err != nil {
		logrus.Errorf("can't do dhcp request")
		return nil, err
	}
	opts := packet.ParseOptions()

	ipnet := net.IPNet{
		IP:   packet.YIAddr(),
		Mask: net.IPMask(opts[dhcp4.OptionSubnetMask]),
	}

	err = fusis_net.AddIp(ipnet.String(), n.Interface)
	if err != nil {
		return nil, err
	}
	return ipnet.IP.String(), nil
}

func (n None) UnsetVip(setResult interface{}) error {
	logrus.Error("Deleting vip")
	return nil
}
