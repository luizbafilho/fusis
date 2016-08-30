package ipam

import (
	"errors"
	"strings"

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

var (
	ErrNoVipAvailable = errors.New("No VIPs available")
)

//Init initilizes ipam module
func New(state *state.State, config *config.BalancerConfig) (Allocator, error) {
	rangeCursor, err := ipaddr.Parse(strings.Join(config.Ipam.Ranges, ","))
	if err != nil {
		return nil, err
	}

	return &Ipam{rangeCursor, state, config}, nil
}

//Allocate allocates a new avaliable ip
func (i *Ipam) AllocateVIP(s *types.Service) error {
	for pos := i.rangeCursor.Next(); pos != nil; pos = i.rangeCursor.Next() {
		//TODO: make it compatible if IPv6
		// Verifies if it is a base IP address. example: 10.0.100.0; 192.168.1.0
		if pos.IP.To4()[3] == 0 {
			pos = i.rangeCursor.Next()
		}

		assigned := i.ipIsAssigned(pos.IP.String(), i.state)

		if !assigned {
			i.rangeCursor.Set(i.rangeCursor.First())
			s.Host = pos.IP.String()
			return nil
		}
	}

	return ErrNoVipAvailable
}

//Release releases a allocated IP
func (i *Ipam) ReleaseVIP(s types.Service) error {
	return nil
}

func (i *Ipam) ipIsAssigned(e string, state state.Store) bool {
	services := state.GetServices()

	for _, a := range services {
		if a.Host == e {
			return true
		}

	}
	return false
}
