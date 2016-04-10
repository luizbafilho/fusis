package engine

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/google/seesaw/ipvs"
)

const (
	NatMode    = ipvs.DFForwardMasq
	TunnelMode = ipvs.DFForwardTunnel
	RouteMode  = ipvs.DFForwardRoute
)

type Service struct {
	Id           string `storm:"id"`
	Name         string `storm:"unique"`
	Host         string
	Port         uint16 `valid:"required"`
	Protocol     string `valid:"required"`
	Scheduler    string `valid:"required"`
	Destinations []Destination
}

type Destination struct {
	Id        string `storm:"id"`
	Name      string `storm:"unique"`
	Host      string `valid:"required"`
	Port      uint16 `valid:"required"`
	Weight    int32
	Mode      string `valid:"required"`
	ServiceId string `storm:"index" valid:"required"`
}

func (svc Service) GetId() string {
	return svc.Name
}

func (dst Destination) GetId() string {
	return dst.Name
}

func stringToIPProto(s string) ipvs.IPProto {
	var value ipvs.IPProto
	if s == "udp" {
		value = syscall.IPPROTO_UDP
	} else {
		value = syscall.IPPROTO_TCP
	}

	return value
}

//MarshalJSON ...
func ipProtoToString(proto ipvs.IPProto) string {
	var value string

	if proto == syscall.IPPROTO_UDP {
		value = "udp"
	} else {
		value = "tcp"
	}

	return value
}

func stringToDestinationFlags(s string) ipvs.DestinationFlags {
	var flag ipvs.DestinationFlags

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
func destinationFlagsToString(flags ipvs.DestinationFlags) string {
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

func (s Service) ToIpvsService() *ipvs.Service {
	destinations := []*ipvs.Destination{}

	for _, dst := range s.Destinations {
		destinations = append(destinations, dst.ToIpvsDestination())
	}

	return &ipvs.Service{
		Address:      net.ParseIP(s.Host),
		Port:         s.Port,
		Protocol:     stringToIPProto(s.Protocol),
		Scheduler:    s.Scheduler,
		Destinations: destinations,
	}
}

func (d Destination) ToIpvsDestination() *ipvs.Destination {
	return &ipvs.Destination{
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

func NewService(s *ipvs.Service) Service {
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

func newDestinationRequest(d *ipvs.Destination) Destination {
	return Destination{
		Host:   d.Address.String(),
		Port:   d.Port,
		Weight: d.Weight,
		Mode:   destinationFlagsToString(d.Flags),
	}
}

func GetServiceFromId(serviceId string) (*Service, error) {
	serviceAttrs := strings.Split(serviceId, "-")

	port, err := strconv.ParseUint(serviceAttrs[1], 10, 16)

	if err != nil {
		return nil, err
	}

	return &Service{
		Host:     serviceAttrs[0],
		Port:     uint16(port),
		Protocol: serviceAttrs[2],
	}, nil
}
