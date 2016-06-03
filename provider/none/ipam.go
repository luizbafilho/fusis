package none

import (
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/mikioh/ipaddr"
)

type Ipam struct {
	rangeCursor *ipaddr.Cursor
	state       ipvs.State
}

//Init initilizes ipam module
func NewIpam(iprange string, state ipvs.State) (*Ipam, error) {
	// var err error
	rangeCursor, err := ipaddr.Parse(iprange)
	if err != nil {
		return nil, err
	}

	return &Ipam{rangeCursor, state}, nil
}

//Allocate allocates a new avaliable ip
func (i *Ipam) Allocate() (string, error) {
	for pos := i.rangeCursor.Next(); pos != nil; pos = i.rangeCursor.Next() {
		assigned, err := i.ipIsAssigned(pos.IP.String())
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

func (i *Ipam) ipIsAssigned(e string) (bool, error) {
	services := i.state.GetServices()

	for _, a := range *services {
		if a.Host == e {
			return true, nil
		}

	}
	return false, nil
}
