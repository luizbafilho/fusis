package ipam

import (
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/state"
	"github.com/mikioh/ipaddr"
)

type Allocator interface {
	AllocateVIP(s *types.Service) error
	ReleaseVIP(s types.Service) error
}

type Ipam struct {
	rangeCursor *ipaddr.Cursor
	state       *state.State
	config      *config.BalancerConfig
}

//Init initilizes ipam module
func New(state *state.State, config *config.BalancerConfig) (Allocator, error) {
	rangeCursor, err := ipaddr.Parse(config.Ipam.Ranges[0])
	if err != nil {
		return nil, err
	}

	return &Ipam{rangeCursor, state, config}, nil
}

//Allocate allocates a new avaliable ip
func (i *Ipam) AllocateVIP(s *types.Service) error {
	for pos := i.rangeCursor.Next(); pos != nil; pos = i.rangeCursor.Next() {
		assigned, err := i.ipIsAssigned(pos.IP.String(), i.state)
		if err != nil {
			return err
		}

		if !assigned {
			i.rangeCursor.Set(i.rangeCursor.First())
			s.Host = pos.IP.String()
			return nil
		}
	}

	return nil
}

//Release releases a allocated IP
func (i *Ipam) ReleaseVIP(s types.Service) error {
	return nil
}

func (i *Ipam) ipIsAssigned(e string, state state.Store) (bool, error) {
	services := state.GetServices()

	for _, a := range services {
		if a.Host == e {
			return true, nil
		}

	}
	return false, nil
}
