package config

import "github.com/luizbafilho/fusis/net"

/* Sample Config

ipam:
	ranges:
		- "192.168.0.0/28"
ports:
	raft: 4382
	serf: 7946
stats:
	interval: 5
	type: "syslog"
	params:
		protocol: "udp"
		host: "logstash_ip_or_domain_address"
		port: "8515"
bgp:
	as: 100
	router-id: "192.168.151.176"
	neighbors:
		-
			address: "192.168.151.178"
			peer-as: 100

*/
type BalancerConfig struct {
	PublicInterface  string `validate:"required"`
	PrivateInterface string

	Name        string `validate:"required"`
	Ports       map[string]int
	Join        []string
	DevMode     bool
	Bootstrap   bool
	DataPath    string
	ClusterMode string `mapstructure:"cluster-mode"` //Defines if balancer is in UNICAST or ANYCAST
	Bgp         Bgp
	Ipam        Ipam
	Metrics     Metrics
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

type Bgp struct {
	As        uint32     `validate:"required"`
	RouterId  string     `validate:"ipv4,required" mapstructure:"router-id"`
	Neighbors []Neighbor `validate:"required,dive"`
}

type Neighbor struct {
	Address string `validate:"ipv4,required"`
	PeerAs  uint32 `validate:"required" mapstructure:"peer-as"`
}

type Ipam struct {
	Ranges []string `validate:"dive,cidrv4"`
}

type Metrics struct {
	Publisher string
	Interval  uint16
	Params    map[string]interface{}
	Extras    map[string]string
}

type Ports struct {
	Raft uint16
	Serf uint16
	Api  uint16
}

func (c *BalancerConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.PublicInterface)
}

func (c *AgentConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}
