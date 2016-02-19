package main

import (
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
	Destinations []*DestinationRequest
}

type DestinationRequest struct {
	Host   net.IP
	Port   uint16
	Weight int32
	mode   DestinationFlags
}

type IPProto ipvs.IPProto

//UnmarshalJSON ...
func (proto *IPProto) UnmarshalJSON(text []byte) error {
	value := strings.ToLower(string(text[1 : len(text)-1])) // Avoid converting the quotes

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

type DestinationFlags uint32

//UnmarshalJSON ...
func (flags *DestinationFlags) UnmarshalJSON(text []byte) error {
	value := strings.ToLower(string(text[1 : len(text)-1])) // Avoid converting the quotes

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
