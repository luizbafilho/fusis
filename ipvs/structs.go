package ipvs

import (
	"encoding/json"
	"errors"
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
	Id           string `storm:"id"`
	Name         string `storm:"unique" valid:"required"`
	Host         string
	Port         uint16 `valid:"required"`
	Protocol     string `valid:"required"`
	Scheduler    string `valid:"required"`
	Destinations []Destination
}

type Destination struct {
	Id        string `storm:"id"`
	Name      string `storm:"unique" valid:"required"`
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

func (s Service) ValidateUniqueness() (bool, error) {
	if s.presentInStore() {
		return false, errors.New("Service found in store")
	}

	if s.presentInKernel() {
		return false, errors.New("Service found in kernel")
	}

	return true, nil
}

func (s Service) presentInStore() bool {
	// svc, _ := Store.GetService(s.Name)
	// if svc != nil {
	// 	return true
	// }

	return false
}

func (s Service) presentInKernel() bool {
	// svc, _ := Kernel.GetService(s.ToIpvsService())
	// if svc != nil {
	// 	return true
	// }

	return false
}

func (d Destination) ValidateUniqueness(svc *Service) (bool, error) {
	if d.presentInStore() {
		return false, errors.New("Destination found in store")
	}

	if d.presentInKernel(svc) {
		return false, errors.New("Destination found in kernel")
	}

	return true, nil
}

func (d Destination) presentInStore() bool {
	// dst, _ := Store.GetDestination(d.Name)
	// if dst != nil {
	// 	return true
	// }

	return false
}

func (d Destination) presentInKernel(svc *Service) bool {
	// dsts, _ := Kernel.GetDestinations(svc.ToIpvsService())
	// for _, dst := range dsts {
	// 	if dst.Equal(*d.ToIpvsDestination()) {
	// 		return true
	// 	}
	// }
	//
	return false
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
