package fusis

import "github.com/luizbafilho/fusis/net"

type Config struct {
	Interface string
}

type BalancerConfig struct {
	Config
}

type AgentConfig struct {
	Config

	Balancer string
	Name     string
	Host     string
	Port     uint16
	Weight   int32
	Mode     string
	Service  string
}

func (c *Config) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}
