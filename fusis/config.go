package fusis

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
)

type Config struct {
	Interface string
}

// It returns the first IP for a given network interface
func (c *AgentConfig) GetIpByInterface() (string, error) {
	fmt.Println("valor dentro GetIpByInterface:", c.Interface)
	i, err := net.InterfaceByName(c.Interface)
	if err != nil {
		log.Errorf("Erro getting IP address: %v", err)
		return "", err
	}

	addrs, err := i.Addrs()
	if err != nil {
		log.Errorf("Erro getting IP address: %v", err)
		return "", err
	}

	addr, ok := addrs[0].(*net.IPNet)
	if !ok {
		log.Errorf("Erro getting IP address: %v", err)
		return "", err
	}

	addrIP := addr.IP
	return addrIP.String(), nil
}

func (c *Config) GetIpByInterface() (string, error) {
	fmt.Println("valor dentro GetIpByInterface:", c.Interface)
	i, err := net.InterfaceByName(c.Interface)
	if err != nil {
		log.Errorf("Erro getting IP address: %v", err)
		return "", err
	}

	addrs, err := i.Addrs()
	if err != nil {
		log.Errorf("Erro getting IP address: %v", err)
		return "", err
	}

	addr, ok := addrs[0].(*net.IPNet)
	if !ok {
		log.Errorf("Erro getting IP address: %v", err)
		return "", err
	}

	addrIP := addr.IP
	return addrIP.String(), nil
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
