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
// "Stats": {
//   "Interval": 5,
//   "Type": "syslog"
//   "params": {
//	   "protocol": "udp",
//     "host": "logstash_ip_or_domain_address",
//     "port": "8515"
//   }
//  }
//}
type Provider struct {
	Type   string
	Params map[string]string
}

type Stats struct {
	Type     string
	Interval uint16
	Params   map[string]string
}

type BalancerConfig struct {
	Interface string

	Name        string
	Bootstrap   bool
	Join        []string
	Provider    Provider
	Stats       Stats
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
