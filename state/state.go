package state

import (
	"sync"

	"github.com/luizbafilho/fusis/types"
)

//go:generate mockery -name=State

type Services []types.Service
type Destinations []types.Destination

type State interface {
	GetServices() []types.Service
	AddService(svc types.Service)

	SetDestinations(dsts Destinations)
	GetDestinations(svc *types.Service) []types.Destination
	AddDestination(dst types.Destination)

	Copy() State
}

// State...
type FusisState struct {
	sync.RWMutex

	services     Services
	destinations Destinations
}

// New creates a new Engine
func New() (State, error) {
	state := &FusisState{
		services:     Services{},
		destinations: Destinations{},
	}

	return state, nil
}

func (s *FusisState) Copy() State {
	s.RLock()
	defer s.RUnlock()

	svcs := Services{}
	dsts := Destinations{}

	copy(svcs, s.services)
	copy(dsts, s.destinations)
	copy := &FusisState{
		services:     svcs,
		destinations: dsts,
	}

	return copy
}

func (s *FusisState) GetServices() []types.Service {
	s.RLock()
	defer s.RUnlock()
	return s.services
}

func (s *FusisState) AddService(svc types.Service) {
	s.Lock()
	defer s.Unlock()
	s.services = append(s.services, svc)
}

func (s *FusisState) SetDestinations(dsts Destinations) {
	s.destinations = dsts
}

func (s *FusisState) GetDestinations(svc *types.Service) []types.Destination {
	s.RLock()
	defer s.RUnlock()
	return s.destinations
}

func (s *FusisState) AddDestination(dst types.Destination) {
	s.Lock()
	defer s.Unlock()
	s.destinations = append(s.destinations, dst)
}
