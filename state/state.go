package state

import (
	"sync"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/store"
)

type State interface {
	GetServices() []types.Service
	GetService(name string) (*types.Service, error)
	AddService(svc types.Service)
	DeleteService(svc *types.Service)

	GetDestination(name string) (*types.Destination, error)
	GetDestinations(svc *types.Service) []types.Destination
	AddDestination(dst types.Destination)
	DeleteDestination(dst *types.Destination)

	Copy() State
}

type Services map[string]types.Service
type Destinations map[string]types.Destination

// State...
type FusisState struct {
	sync.RWMutex

	store store.Store

	services     Services
	destinations Destinations

	changesCh chan bool
}

// New creates a new Engine
func New(store store.Store, changesCh chan bool) (State, error) {
	state := &FusisState{
		services:     make(Services),
		destinations: make(Destinations),
		changesCh:    changesCh,
		store:        store,
	}

	go state.handleServicesChange()
	go state.handleDestinationsChange()

	return state, nil
}

func (s *FusisState) Copy() State {
	s.RLock()
	defer s.RUnlock()

	copy := &FusisState{
		services:     make(Services),
		destinations: make(Destinations),
	}

	for _, svc := range s.services {
		copy.AddService(svc)
	}

	for _, dst := range s.destinations {
		copy.AddDestination(dst)
	}

	return copy
}

func (s *FusisState) watchStore() {
}

func (s *FusisState) handleServicesChange() {
	updateCh := make(chan []types.Service)
	s.store.SubscribeServices(updateCh)

	for {
		svcs := <-updateCh
		s.updateServices(svcs)
		s.changesCh <- true
	}
}

func (s *FusisState) handleDestinationsChange() {
	updateCh := make(chan []types.Destination)
	s.store.SubscribeDestinations(updateCh)

	for {
		dsts := <-updateCh
		s.updateDestinations(dsts)
		s.changesCh <- true
	}
}

func (s *FusisState) GetServices() []types.Service {
	s.RLock()
	defer s.RUnlock()

	services := []types.Service{}
	for _, v := range s.services {
		services = append(services, v)
	}
	return services
}

func (s *FusisState) GetService(name string) (*types.Service, error) {
	s.RLock()
	defer s.RUnlock()

	svc, ok := s.services[name]
	if !ok {
		return nil, types.ErrServiceNotFound
	}

	return &svc, nil
}

func (s *FusisState) updateServices(svcs []types.Service) {
	s.Lock()
	s.services = Services{}
	s.Unlock()

	for _, svc := range svcs {
		s.AddService(svc)
	}
}

func (s *FusisState) AddService(svc types.Service) {
	s.Lock()
	defer s.Unlock()

	s.services[svc.GetId()] = svc
}

func (s *FusisState) DeleteService(svc *types.Service) {
	s.Lock()
	defer s.Unlock()

	delete(s.services, svc.GetId())
}

func (s *FusisState) GetDestinations(svc *types.Service) []types.Destination {
	s.RLock()
	defer s.RUnlock()

	dsts := []types.Destination{}
	for _, d := range s.destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}

	return dsts
}

func (s *FusisState) GetDestination(name string) (*types.Destination, error) {
	s.RLock()
	defer s.RUnlock()

	dst := s.destinations[name]
	if dst.Name == "" {
		return nil, types.ErrDestinationNotFound
	}
	return &dst, nil
}

func (s *FusisState) AddDestination(dst types.Destination) {
	s.Lock()
	defer s.Unlock()

	s.destinations[dst.GetId()] = dst
}

func (s *FusisState) DeleteDestination(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	delete(s.destinations, dst.GetId())
}

func (s *FusisState) updateDestinations(dsts []types.Destination) {
	s.Lock()
	s.destinations = Destinations{}
	s.Unlock()
	for _, dst := range dsts {
		s.AddDestination(dst)
	}
}
