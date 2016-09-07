package ipvs

import (
	"net"
	"syscall"

	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api/types"
	cipvs "github.com/qmsk/clusterf/ipvs"
)

const (
	NatMode    = gipvs.DFForwardMasq
	TunnelMode = gipvs.DFForwardTunnel
	RouteMode  = gipvs.DFForwardRoute
)

func stringToIPProto(s string) cipvs.Protocol {
	var value cipvs.Protocol
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

// func ToIpvsService(s *types.Service) *gipvs.Service {
// 	// destinations := []*gipvs.Destination{}
// 	// for _, dest := range s.Destinations {
// 	// 	destinations = append(destinations, toIpvsDestination(&dest))
// 	// }
// 	//
// 	// service := &gipvs.Service{
// 	// 	Address:      net.ParseIP(s.Host),
// 	// 	Port:         s.Port,
// 	// 	Protocol:     stringToIPProto(s.Protocol),
// 	// 	Scheduler:    s.Scheduler,
// 	// 	Destinations: destinations,
// 	// }
// 	//
// 	// if s.Persistent > 0 {
// 	// 	service.Flags = gipvs.SFPersistent
// 	// 	service.Timeout = s.Persistent
// 	// }
// 	//
// 	// return service
// 	return nil
// }

func toIpvsDestination(d *types.Destination) *gipvs.Destination {
	return &gipvs.Destination{
		Address: net.ParseIP(d.Host),
		Port:    d.Port,
		Weight:  d.Weight,
		Flags:   stringToDestinationFlags(d.Mode),
	}
}

func FromService(s *gipvs.Service) types.Service {
	// destinations := []types.Destination{}
	// for _, dst := range s.Destinations {
	// 	destinations = append(destinations, fromDestination(dst))
	// }
	//
	// return types.Service{
	// 	Host:         s.Address.String(),
	// 	Port:         s.Port,
	// 	Protocol:     ipProtoToString(s.Protocol),
	// 	Scheduler:    s.Scheduler,
	// 	Destinations: destinations,
	// }
	return types.Service{}
}

// func fromDestination(d *gipvs.Destination) types.Destination {
// 	return types.Destination{
// 		Host:   d.Address.String(),
// 		Port:   d.Port,
// 		Weight: d.Weight,
// 		Mode:   destinationFlagsToString(d.Flags),
// 	}
// }

// New IPVS lib converters

// type Service struct {
// 	// id
// 	Af       Af
// 	Protocol Protocol
// 	Addr     net.IP
// 	Port     uint16
// 	FwMark   uint32
//
// 	// params
// 	SchedName string
// 	Flags     Flags
// 	Timeout   uint32
// 	Netmask   uint32
// }

func ToIpvsService(s *types.Service) cipvs.Service {
	// destinations := []*cipvs.Destination{}
	// for _, dest := range s.Destinations {
	// 	destinations = append(destinations, toIpvsDestination(&dest))
	// }

	service := cipvs.Service{
		Af:        syscall.AF_INET,
		Addr:      net.ParseIP(s.Host),
		Port:      s.Port,
		Protocol:  stringToIPProto(s.Protocol),
		SchedName: s.Scheduler,
	}

	if s.Persistent > 0 {
		service.Flags = cipvs.Flags{cipvs.IP_VS_SVC_F_PERSISTENT, 0}
		service.Timeout = s.Persistent
	}

	return service
}

// type Dest struct {
// 	// id
// 	// TODO: IPVS_DEST_ATTR_ADDR_FAMILY
// 	Addr net.IP
// 	Port uint16
//
// 	// params
// 	FwdMethod FwdMethod
// 	Weight    uint32
// 	UThresh   uint32
// 	LThresh   uint32
//
// 	// info
// 	ActiveConns  uint32
// 	InactConns   uint32
// 	PersistConns uint32
// 	Stats        Stats
// }
func toIpvsDest(d *types.Destination) cipvs.Dest {
	fwMode, _ := cipvs.ParseFwdMethod(d.Mode)
	return cipvs.Dest{
		Addr:      net.ParseIP(d.Host),
		Port:      d.Port,
		Weight:    uint32(d.Weight),
		FwdMethod: fwMode,
	}

}

func fromService(s cipvs.Service) types.Service {
	// destinations := []types.Destination{}
	// for _, dst := range s.Destinations {
	// 	destinations = append(destinations, fromDestination(dst))
	// }

	return types.Service{
		Host:      s.Addr.String(),
		Port:      s.Port,
		Protocol:  s.Protocol.String(),
		Scheduler: s.SchedName,
		// Destinations: destinations,
	}
}

func fromDestination(d *cipvs.Dest) types.Destination {
	return types.Destination{
		Host:   d.Addr.String(),
		Port:   d.Port,
		Weight: int32(d.Weight),
		Mode:   d.FwdMethod.String(),
	}
}
