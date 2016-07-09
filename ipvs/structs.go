package ipvs

import (
	"net"
	"syscall"

	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api/types"
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
	for _, dest := range s.Destinations {
		destinations = append(destinations, toIpvsDestination(&dest))
	}

	return &gipvs.Service{
		Address:      net.ParseIP(s.Host),
		Port:         s.Port,
		Protocol:     stringToIPProto(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
	}
}

func toIpvsDestination(d *types.Destination) *gipvs.Destination {
	return &gipvs.Destination{
		Address: net.ParseIP(d.Host),
		Port:    d.Port,
		Weight:  d.Weight,
		Flags:   stringToDestinationFlags(d.Mode),
	}
}

func getServiceStats(s *gipvs.Service) *types.ServiceStats {

	return &types.ServiceStats{
		Connections: s.Statistics.Connections,
		PacketsIn:   s.Statistics.PacketsIn,
		PacketsOut:  s.Statistics.PacketsOut,
		BytesIn:     s.Statistics.BytesIn,
		BytesOut:    s.Statistics.BytesOut,
		CPS:         s.Statistics.CPS,
		PPSIn:       s.Statistics.PPSIn,
		PPSOut:      s.Statistics.PPSOut,
		BPSIn:       s.Statistics.BPSIn,
		BPSOut:      s.Statistics.BPSOut,
	}
}

func getDestinationStats(d *gipvs.Destination) *types.DestinationStats {

	return &types.DestinationStats{
		ActiveConns:   d.Statistics.ActiveConns,
		InactiveConns: d.Statistics.InactiveConns,
		PersistConns:  d.Statistics.PersistConns,
	}
}

func FromService(s *gipvs.Service) types.Service {
	destinations := []types.Destination{}
	for _, dst := range s.Destinations {
		destinations = append(destinations, fromDestination(dst))
	}

	return types.Service{
		Host:         s.Address.String(),
		Port:         s.Port,
		Protocol:     ipProtoToString(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
		Stats:        getServiceStats(s),
	}
}

func fromDestination(d *gipvs.Destination) types.Destination {
	return types.Destination{
		Host:   d.Address.String(),
		Port:   d.Port,
		Weight: d.Weight,
		Mode:   destinationFlagsToString(d.Flags),
		Stats:  getDestinationStats(d),
	}
}
