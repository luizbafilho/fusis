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

type Config struct {
	Interface string
}

type BalancerConfig struct {
	Config

	Single     bool
	Join       string
	Provider   Provider
	ConfigPath string
	RaftPort   int
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

var Balancer BalancerConfig
