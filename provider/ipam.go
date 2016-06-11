package provider

import (
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/mikioh/ipaddr"
)

type Ipam struct {
	rangeCursor *ipaddr.Cursor
}

//Init initilizes ipam module
func NewIpam(iprange string) (*Ipam, error) {
	// var err error
	rangeCursor, err := ipaddr.Parse(iprange)
	if err != nil {
		return nil, err
	}

	return &Ipam{rangeCursor}, nil
}

//Allocate allocates a new avaliable ip
func (i *Ipam) Allocate(state ipvs.State) (string, error) {
	for pos := i.rangeCursor.Next(); pos != nil; pos = i.rangeCursor.Next() {
		assigned, err := i.ipIsAssigned(pos.IP.String(), state)
		if err != nil {
			return "", err
		}

		if !assigned {
			i.rangeCursor.Set(i.rangeCursor.First())
			return pos.IP.String(), nil
		}
	}

	return "", nil
}

//Release releases a allocated IP
func (i *Ipam) Release(allocIP string) {}

func (i *Ipam) ipIsAssigned(e string, state ipvs.State) (bool, error) {
	services := state.GetServices()

	for _, a := range *services {
		if a.Host == e {
			return true, nil
		}

	}
	return false, nil
}
