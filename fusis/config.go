package fusis

import "github.com/luizbafilho/fusis/net"

type BalancerConfig struct {
	Interface string
}

type AgentConfig struct {
	Balancer  string
	Name      string
	Host      string
	Port      uint16
	Weight    int32
	Mode      string
	Service   string
	Interface string
}

// It returns the first IP for a given network interface
func (c *AgentConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}

func (c *BalancerConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}
