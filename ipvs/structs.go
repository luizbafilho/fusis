package ipvs

import (
	"net"
	"syscall"

	"github.com/docker/libnetwork/ipvs"
	"github.com/luizbafilho/fusis/types"
	"github.com/mqliang/libipvs"
)

func stringToIPProto(s string) uint16 {
	var value uint16
	if s == "udp" {
		value = syscall.IPPROTO_UDP
	} else {
		value = syscall.IPPROTO_TCP
	}

	return value
}

//MarshalJSON ...
func ipProtoToString(proto uint16) string {
	var value string

	if proto == syscall.IPPROTO_UDP {
		value = "udp"
	} else {
		value = "tcp"
	}

	return value
}

func stringToDestinationFlags(s string) uint32 {
	var flag uint32

	switch s {
	case "nat":
		flag = ipvs.ConnectionFlagMasq
	case "tunnel":
		flag = ipvs.ConnectionFlagTunnel
	default:
		// Default is Direct Routing
		flag = libipvs.IP_VS_CONN_F_DROUTE
	}

	return flag
}

//MarshalJSON ...
func destinationFlagsToString(fwdMethod uint32) string {
	var value string

	switch fwdMethod {
	case ipvs.ConnectionFlagMasq:
		value = "nat"
		// *flags =
	case ipvs.ConnectionFlagTunnel:
		value = "tunnel"
	default:
		// Default is Direct Routing
		value = "route"
	}

	return value
}

func ToIpvsService(s *types.Service) *ipvs.Service {
	service := &ipvs.Service{
		AddressFamily: syscall.AF_INET,
		Address:       net.ParseIP(s.Address),
		Protocol:      stringToIPProto(s.Protocol),
		Port:          s.Port,
		SchedName:     s.Scheduler,
	}

	if s.Persistent > 0 {
		// defining the IP_VS_SVC_F_PERSISTENT flag
		service.Flags = 1
		service.Timeout = s.Persistent
	}

	return service
}

func ToIpvsDestination(d *types.Destination) *ipvs.Destination {
	return &ipvs.Destination{
		AddressFamily:   syscall.AF_INET,
		Address:         net.ParseIP(d.Address),
		Port:            d.Port,
		Weight:          int(d.Weight),
		ConnectionFlags: stringToDestinationFlags(d.Mode),
	}
}

func FromService(s *ipvs.Service) types.Service {
	return types.Service{
		Address:   s.Address.String(),
		Port:      s.Port,
		Protocol:  ipProtoToString(s.Protocol),
		Scheduler: s.SchedName,
	}
}

func fromDestination(d *ipvs.Destination) types.Destination {
	return types.Destination{
		Address: d.Address.String(),
		Port:    d.Port,
		Weight:  int32(d.Weight),
		Mode:    destinationFlagsToString(d.ConnectionFlags),
	}
}
