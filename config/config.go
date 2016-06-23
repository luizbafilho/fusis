package config

import "github.com/luizbafilho/fusis/net"

// {
// 	"provider": {
// 		"type": "cloudstack",
// 		"params": {
// 			"apiKey": "seila",
// 			"secretKey": "testando",
//		  "vipRange":"192.168.0.1/24"
// 		}
// 	}
// }
type Provider struct {
	Type   string
	Params map[string]string
}

type BalancerConfig struct {
	Interface string

    Name        string
    Bootstrap   bool
    Join        []string
    Provider    Provider
    ConfigPath  string
    Ports       map[string]int
    DevMode     bool
    LogInterval uint16
}

type AgentConfig struct {
	Interface string

	Balancer string
	Name     string
	Host     string
	Port     uint16
	Weight   int32
	Mode     string
	Service  string
}

func (c *BalancerConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}

func (c *AgentConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}
