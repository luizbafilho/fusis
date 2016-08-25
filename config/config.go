package config

import "github.com/luizbafilho/fusis/net"

// provider:
// 	type: "none"
// 	params:
// 		interface: "eth0"
// 		vip-range: "192.168.0.0/28"
// ports:
// 	raft: 4382
// 	serf: 7946
// stats:
// 	interval: 5
// 	type: "syslog"
// 	params:
// 		protocol: "udp"
// 		host: "logstash_ip_or_domain_address"
// 		port: "8515"
// bgp:
// 	as: 100
// 	router-id: "192.168.151.176"
// 	neighbors:
// 		-
// 			address: "192.168.151.178"
// 			peer-as: 100
type Provider struct {
	Type   string
	Params map[string]string
}

type Stats struct {
	Type     string
	Interval uint16
	Params   map[string]string
}

type Bgp struct {
	As        uint32     `valid:"required"`
	RouterId  string     `valid:"ipv4,required" mapstructure:"router-id"`
	Neighbors []Neighbor `valid:"required"`
}

type Neighbor struct {
	Address string `valid:"ipv4,required"`
	PeerAs  uint32 `valid:"required" mapstructure:"peer-as"`
}

type BalancerConfig struct {
	PublicInterface  string
	PrivateInterface string

	Name        string
	Bootstrap   bool
	Join        []string
	Provider    Provider
	Stats       Stats
	ConfigPath  string
	Ports       map[string]int
	DevMode     bool
	ClusterMode string `mapstructure:"cluster-mode"` //Defines if balancer is in UNICAST or ANYCAST
	Bgp         Bgp
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
	return net.GetIpByInterface(c.PublicInterface)
}

func (c *AgentConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}
