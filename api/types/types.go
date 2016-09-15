package types

import (
	"errors"
	"fmt"
)

var (
	ErrServiceNotFound          error = ErrNotFound("service not found")
	ErrDestinationNotFound      error = ErrNotFound("destination not found")
	ErrServiceAlreadyExists           = errors.New("service already exists")
	ErrDestinationAlreadyExists       = errors.New("destination already exists")
)

var (
	NAT    = "nat"
	ROUTE  = "route"
	TUNNEL = "tunnel"
)

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return string(e)
}

type Service struct {
	Name       string `valid:"required"`
	Host       string
	Port       uint16 `valid:"required"`
	Protocol   string `valid:"required"`
	Scheduler  string `valid:"required"`
	Mode       string `valid:"required"`
	Persistent uint32
}

type Destination struct {
	Name      string `valid:"required"`
	Address   string `valid:"required"`
	Port      uint16 `valid:"required"`
	Weight    int32
	Mode      string `valid:"required"`
	ServiceId string `valid:"required"`
}

func (svc Service) GetId() string {
	return svc.Name
}

func (svc Service) IsNat() bool {
	return svc.Mode == NAT
}

func (dst Destination) GetId() string {
	return dst.Name
}

func (svc Service) KernelKey() string {
	return fmt.Sprintf("%s-%d-%s", svc.Host, svc.Port, svc.Protocol)
}

func (dst Destination) KernelKey() string {
	return fmt.Sprintf("%s-%d", dst.Address, dst.Port)
}

type DestinationList []Destination

func (l DestinationList) Len() int      { return len(l) }
func (l DestinationList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l DestinationList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}
