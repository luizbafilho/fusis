package state

import (
	"sync"

	"github.com/luizbafilho/fusis/types"
)

//go:generate mockery -name=State

type Services []types.Service
type Destinations map[string]types.Destination
type Checks []types.CheckSpec

type State interface {
	GetServices() []types.Service
	AddService(svc types.Service)

	// SetDestinations(dsts Destinations)
	GetDestinations(svc *types.Service) []types.Destination
	AddDestination(dst types.Destination)
	DeleteDestination(dst types.Destination)

	AddCheck(c types.CheckSpec)
	GetChecks() []types.CheckSpec
	SetChecks(checks Checks)
	Copy() State
}

// State...
type FusisState struct {
	sync.RWMutex

	services     Services
	destinations Destinations
	checks       Checks
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

	new := &FusisState{
		services:     Services{},
		destinations: Destinations{},
	}

	copy(new.services, s.services)

	for _, dst := range s.destinations {
		new.AddDestination(dst)
	}

	return new
}

func (s *FusisState) AddCheck(check types.CheckSpec) {
	s.Lock()
	defer s.Unlock()
	s.checks = append(s.checks, check)
}

func (s *FusisState) GetChecks() []types.CheckSpec {
	s.RLock()
	defer s.RUnlock()
	return s.checks
}

func (s *FusisState) SetChecks(checks Checks) {
	s.Lock()
	defer s.Unlock()
	s.checks = checks
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

func (s *FusisState) GetDestinations(svc *types.Service) []types.Destination {
	s.RLock()
	defer s.RUnlock()

	dsts := []types.Destination{}
	if svc == nil {
		for _, d := range s.destinations {
			dsts = append(dsts, d)
		}

		return dsts
	}

	for _, d := range s.destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}

	return dsts
}

func (s *FusisState) AddDestination(dst types.Destination) {
	s.Lock()
	defer s.Unlock()

	s.destinations[dst.GetId()] = dst
}

func (s *FusisState) DeleteDestination(dst types.Destination) {
	delete(s.destinations, dst.GetId())
}
