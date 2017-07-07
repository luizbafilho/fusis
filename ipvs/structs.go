package ipvs

import (
	"net"
	"syscall"

	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/types"
	"github.com/mqliang/libipvs"
)

const (
	NatMode    = gipvs.DFForwardMasq
	TunnelMode = gipvs.DFForwardTunnel
	RouteMode  = gipvs.DFForwardRoute
)

func stringToIPProto(s string) libipvs.Protocol {
	var value libipvs.Protocol
	if s == "udp" {
		value = syscall.IPPROTO_UDP
	} else {
		value = syscall.IPPROTO_TCP
	}

	return value
}

//MarshalJSON ...
func ipProtoToString(proto libipvs.Protocol) string {
	var value string

	if proto == syscall.IPPROTO_UDP {
		value = "udp"
	} else {
		value = "tcp"
	}

	return value
}

func stringToDestinationFlags(s string) libipvs.FwdMethod {
	var flag libipvs.FwdMethod

	switch s {
	case "nat":
		flag = libipvs.IP_VS_CONN_F_MASQ
	case "tunnel":
		flag = libipvs.IP_VS_CONN_F_TUNNEL
	default:
		// Default is Direct Routing
		flag = libipvs.IP_VS_CONN_F_DROUTE
	}

	return flag
}

//MarshalJSON ...
func destinationFlagsToString(fwdMethod libipvs.FwdMethod) string {
	var value string

	switch fwdMethod {
	case libipvs.IP_VS_CONN_F_MASQ:
		value = "nat"
		// *flags =
	case libipvs.IP_VS_CONN_F_TUNNEL:
		value = "tunnel"
	default:
		// Default is Direct Routing
		value = "route"
	}

	return value
}

func ToIpvsService(s *types.Service) *libipvs.Service {
	service := &libipvs.Service{
		AddressFamily: syscall.AF_INET,
		Address:       net.ParseIP(s.Address),
		Protocol:      stringToIPProto(s.Protocol),
		Port:          s.Port,
		SchedName:     s.Scheduler,
	}

	if s.Persistent > 0 {
		service.Flags = libipvs.Flags{Flags: libipvs.IP_VS_SVC_F_PERSISTENT}
		service.Timeout = s.Persistent
	}

	return service
}

func ToIpvsDestination(d *types.Destination) *libipvs.Destination {
	return &libipvs.Destination{
		AddressFamily: syscall.AF_INET,
		Address:       net.ParseIP(d.Address),
		Port:          d.Port,
		Weight:        uint32(d.Weight),
		FwdMethod:     stringToDestinationFlags(d.Mode),
	}
}

func FromService(s *libipvs.Service) types.Service {
	return types.Service{
		Address:   s.Address.String(),
		Port:      s.Port,
		Protocol:  ipProtoToString(s.Protocol),
		Scheduler: s.SchedName,
	}
}

func fromDestination(d *libipvs.Destination) types.Destination {
	return types.Destination{
		Address: d.Address.String(),
		Port:    d.Port,
		Weight:  int32(d.Weight),
		Mode:    destinationFlagsToString(d.FwdMethod),
	}
}
