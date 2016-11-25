package config

import (
	"github.com/hashicorp/logutils"
	"github.com/luizbafilho/fusis/net"
)

var (
	LOG_LEVELS = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"}
)

type BalancerConfig struct {
	Interfaces

	Name        string         `mapstructure:"name" validate:"required"`
	Ports       map[string]int `mapstructure:"ports"`
	Join        []string       `mapstructure:"join"`
	DevMode     bool           `mapstructure:"dev-mode"`
	Bootstrap   bool           `mapstructure:"bootstrap"`
	LogLevel    string         `mapstructure:"log-level"`
	DataPath    string         `mapstructure:"data-path"`
	ClusterMode string         `mapstructure:"cluster-mode"` //Defines if balancer is in UNICAST or ANYCAST
	Bgp
	Ipam
	Metrics

	StoreAddress string        `mapstructure:"store-address"`
}

type AgentConfig struct {
	Interface string

	Balancer string
	Name     string
	Address  string
	Port     uint16
	Weight   int32
	Mode     string
	Service  string
}

type Interfaces struct {
	Inbound  string `validate:"required"`
	Outbound string `validate:"required"`
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
	return net.GetIpByInterface(c.Interfaces.Inbound)
}

func (c *AgentConfig) GetIpByInterface() (string, error) {
	return net.GetIpByInterface(c.Interface)
}
