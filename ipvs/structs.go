package ipvs

import (
	"net"
	"syscall"

	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/types"
)

const (
	NatMode    = gipvs.DFForwardMasq
	TunnelMode = gipvs.DFForwardTunnel
	RouteMode  = gipvs.DFForwardRoute
)

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

func ToIpvsService(s *types.Service) *gipvs.Service {
	destinations := []*gipvs.Destination{}
	// for _, dest := range s.Destinations {
	// 	destinations = append(destinations, ToIpvsDestination(&dest))
	// }

	service := &gipvs.Service{
		Address:      net.ParseIP(s.Address),
		Port:         s.Port,
		Protocol:     stringToIPProto(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
	}

	if s.Persistent > 0 {
		service.Flags = gipvs.SFPersistent
		service.Timeout = s.Persistent
	}

	return service
}

func ToIpvsDestination(d *types.Destination) *gipvs.Destination {
	return &gipvs.Destination{
		Address: net.ParseIP(d.Address),
		Port:    d.Port,
		Weight:  d.Weight,
		Flags:   stringToDestinationFlags(d.Mode),
	}
}

func FromService(s *gipvs.Service) types.Service {
	destinations := []types.Destination{}
	for _, dst := range s.Destinations {
		destinations = append(destinations, fromDestination(dst))
	}

	return types.Service{
		Address:   s.Address.String(),
		Port:      s.Port,
		Protocol:  ipProtoToString(s.Protocol),
		Scheduler: s.Scheduler,
		// Destinations: destinations,
	}
}

func fromDestination(d *gipvs.Destination) types.Destination {
	return types.Destination{
		Address: d.Address.String(),
		Port:    d.Port,
		Weight:  d.Weight,
		Mode:    destinationFlagsToString(d.Flags),
	}
}
