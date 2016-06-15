package ipvs

import (
	"encoding/json"
	"net"
	"syscall"

	gipvs "github.com/google/seesaw/ipvs"
)

const (
	NatMode    = gipvs.DFForwardMasq
	TunnelMode = gipvs.DFForwardTunnel
	RouteMode  = gipvs.DFForwardRoute
)

type Service struct {
	Name         string `valid:"required"`
	Host         string
	Port         uint16 `valid:"required"`
	Protocol     string `valid:"required"`
	Scheduler    string `valid:"required"`
	Destinations []Destination
}

type Destination struct {
	Name      string `valid:"required"`
	Host      string `valid:"required"`
	Port      uint16 `valid:"required"`
	Weight    int32
	Mode      string `valid:"required"`
	ServiceId string `valid:"required"`
}

func (svc Service) GetId() string {
	return svc.Name
}

func (dst Destination) GetId() string {
	return dst.Name
}

func stringToIPProto(s string) gipvs.IPProto {
	var value gipvs.IPProto
	if s == "udp" {
		value = syscall.IPPROTO_UDP
	} else {
		value = syscall.IPPROTO_TCP
	}

	return value
}

//MarshalJSON ...
func ipProtoToString(proto gipvs.IPProto) string {
	var value string

	if proto == syscall.IPPROTO_UDP {
		value = "udp"
	} else {
		value = "tcp"
	}

	return value
}

func stringToDestinationFlags(s string) gipvs.DestinationFlags {
	var flag gipvs.DestinationFlags

	switch s {
	case "nat":
		flag = NatMode
	case "tunnel":
		flag = TunnelMode
	default:
		// Default is Direct Routing
		flag = RouteMode
	}

	return flag
}

//MarshalJSON ...
func destinationFlagsToString(flags gipvs.DestinationFlags) string {
	var value string

	switch flags {
	case NatMode:
		value = "nat"
		// *flags =
	case TunnelMode:
		value = "tunnel"
	default:
		// Default is Direct Routing
		value = "route"
	}

	return value
}

func (s Service) ToIpvsService() *gipvs.Service {
	destinations := []*gipvs.Destination{}

	return &gipvs.Service{
		Address:      net.ParseIP(s.Host),
		Port:         s.Port,
		Protocol:     stringToIPProto(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
	}
}

func (d Destination) ToIpvsDestination() *gipvs.Destination {
	return &gipvs.Destination{
		Address: net.ParseIP(d.Host),
		Port:    d.Port,
		Weight:  d.Weight,
		Flags:   stringToDestinationFlags(d.Mode),
	}
}

func (s Service) ToJson() ([]byte, error) {
	return json.Marshal(s)
}

func (d Destination) ToJson() ([]byte, error) {
	return json.Marshal(d)
}

func NewService(s *gipvs.Service) Service {
	destinations := []Destination{}

	for _, dst := range s.Destinations {
		destinations = append(destinations, newDestinationRequest(dst))
	}

	return Service{
		Host:         s.Address.String(),
		Port:         s.Port,
		Protocol:     ipProtoToString(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
	}
}

func newDestinationRequest(d *gipvs.Destination) Destination {
	return Destination{
		Host:   d.Address.String(),
		Port:   d.Port,
		Weight: d.Weight,
		Mode:   destinationFlagsToString(d.Flags),
	}
}
