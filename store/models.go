package store

import (
	"encoding/json"
	"net"
	"strings"
	"syscall"

	"github.com/luizbafilho/janus/ipvs"
)

type ServiceRequest struct {
	Host         net.IP
	Port         uint16
	Protocol     IPProto
	Scheduler    string
	Destinations []DestinationRequest
}

type DestinationRequest struct {
	Host   net.IP
	Port   uint16
	Weight int32
	Mode   DestinationFlags
}

type IPProto ipvs.IPProto

//UnmarshalJSON ...
func (proto *IPProto) UnmarshalJSON(text []byte) error {
	value := strings.ToLower(strings.Trim(string(text), "\"")) // Avoid converting the quotes

	if value == "udp" {
		*proto = syscall.IPPROTO_UDP
	} else {
		*proto = syscall.IPPROTO_TCP
	}

	return nil
}

const (
	NatMode    = DestinationFlags(ipvs.DFForwardMasq)
	TunnelMode = DestinationFlags(ipvs.DFForwardTunnel)
	RouteMode  = DestinationFlags(ipvs.DFForwardRoute)
)

//MarshalJSON ...
func (proto IPProto) MarshalJSON() ([]byte, error) {
	var value string

	if proto == syscall.IPPROTO_UDP {
		value = "udp"
	} else {
		value = "tcp"
	}

	return json.Marshal(value)
}

func (proto IPProto) String() string {
	if proto == syscall.IPPROTO_UDP {
		return "udp"
	} else {
		return "tcp"
	}
}

type DestinationFlags uint32

//UnmarshalJSON ...
func (flags *DestinationFlags) UnmarshalJSON(text []byte) error {
	value := strings.ToLower(strings.Trim(string(text), "\"")) // Avoid converting the quotes

	switch value {
	case "nat":
		*flags = NatMode
	case "tunnel":
		*flags = TunnelMode
	default:
		// Default is Direct Routing
		*flags = RouteMode
	}
	return nil
}

//MarshalJSON ...
func (flags DestinationFlags) MarshalJSON() ([]byte, error) {
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

	return json.Marshal(value)
}

func (s ServiceRequest) ToIpvsService() ipvs.Service {
	destinations := []*ipvs.Destination{}

	for _, dst := range s.Destinations {
		destinations = append(destinations, dst.ToIpvsDestination())
	}

	return ipvs.Service{
		Address:      s.Host,
		Port:         s.Port,
		Protocol:     ipvs.IPProto(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
	}
}

func (d DestinationRequest) ToIpvsDestination() *ipvs.Destination {
	return &ipvs.Destination{
		Address: d.Host,
		Port:    d.Port,
		Weight:  d.Weight,
		Flags:   ipvs.DestinationFlags(d.Mode),
	}
}

func NewServiceRequest(s *ipvs.Service) ServiceRequest {
	destinations := []DestinationRequest{}

	for _, dst := range s.Destinations {
		destinations = append(destinations, newDestinationRequest(dst))
	}

	return ServiceRequest{
		Host:         s.Address,
		Port:         s.Port,
		Protocol:     IPProto(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
	}
}

func newDestinationRequest(d *ipvs.Destination) DestinationRequest {
	return DestinationRequest{
		Host:   d.Address,
		Port:   d.Port,
		Weight: d.Weight,
		Mode:   DestinationFlags(d.Flags),
	}
}
