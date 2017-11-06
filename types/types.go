package types

import (
	"fmt"
	"time"
)

var (
	NAT    = "nat"
	ROUTE  = "route"
	TUNNEL = "tunnel"
)

var (
	Protocols  = []string{"tcp", "udp"}
	Schedulers = []string{"rr", "wrr", "lc"}
)

var (
	ErrServiceNotFound     = ErrNotFound("service not found")
	ErrDestinationNotFound = ErrNotFound("destination not found")
	ErrCheckNotFound       = ErrNotFound("check not found")
	ErrServiceConflict     = ErrConflict("service already exists")
	ErrDestinationConflict = ErrConflict("destination already exists")
)

type ErrConflict string

func (e ErrConflict) Error() string {
	return string(e)
}

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return string(e)
}

type ErrValidation struct {
	Type   string
	Errors map[string]string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("Validation failed. %#v", e)
}

type Service struct {
	Name       string `validate:"required"`
	Address    string
	Port       uint16 `validate:"gte=1,lte=47808,required"`
	Protocol   string `validate:"protocols,required"`
	Scheduler  string `validate:"schedulers,required"`
	Mode       string `validate:"required"`
	Persistent uint32
}

func (svc Service) GetId() string {
	return svc.Name
}

func (svc Service) IsNat() bool {
	return svc.Mode == NAT
}

func (svc Service) IpvsId() string {
	return fmt.Sprintf("%s-%d-%s", svc.Address, svc.Port, svc.Protocol)
}

func (s Service) Equal(svc Service) bool {
	return s.Name == svc.Name || (s.Address == svc.Address && s.Port == svc.Port && s.Protocol == svc.Protocol)
}

type Destination struct {
	Name      string `validate:"required"`
	Address   string `validate:"required"`
	Port      uint16 `validate:"gte=1,lte=47808,required"`
	Weight    int32
	Mode      string `validate:"required"`
	ServiceId string `validate:"required"`
}

func (dst Destination) GetId() string {
	return dst.Name
}

func (dst Destination) IpvsId() string {
	return fmt.Sprintf("%s-%d", dst.Address, dst.Port)
}

func (d Destination) Equal(dst Destination) bool {
	return d.Name == dst.Name || (d.Address == dst.Address && d.Port == dst.Port)
}

type DestinationList []Destination

func (l DestinationList) Len() int      { return len(l) }
func (l DestinationList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l DestinationList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

type CheckSpec struct {
	ServiceID string
	Type      string

	HttpPath string
	Script   string
	Shell    string

	Interval time.Duration
	Timeout  time.Duration
}
