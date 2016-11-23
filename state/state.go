package state

import (
	"sync"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
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

	ChangesCh() chan bool
}

// State...
type FusisState struct {
	sync.Mutex

	services     map[string]types.Service
	destinations map[string]types.Destination

	changesCh chan bool

	store store.Store
}

type Services map[string]types.Service
type Destinations map[string]types.Destination

// New creates a new Engine
func New(store store.Store, config *config.BalancerConfig) (State, error) {
	state := &FusisState{
		services:     make(Services),
		destinations: make(Destinations),
		changesCh:    make(chan bool),
		store:        store,
	}

	state.watchStore()

	return state, nil
}

func (s *FusisState) ChangesCh() chan bool {
	return s.changesCh
}

func (s *FusisState) watchStore() {
	go s.handleServicesChange()
	go s.handleDestinationsChange()
}

func (s *FusisState) handleServicesChange() {
	updateCh := make(chan []types.Service)
	s.store.SubscribeServices(updateCh)

	for {
		svcs := <-updateCh

		s.services = Services{}
		for _, svc := range svcs {
			s.AddService(svc)
		}

		s.changesCh <- true
	}
}

func (s *FusisState) handleDestinationsChange() {
	updateCh := make(chan []types.Destination)
	s.store.SubscribeDestinations(updateCh)

	for {
		dsts := <-updateCh
		s.destinations = Destinations{}
		for _, dst := range dsts {
			s.AddDestination(dst)
		}

		s.changesCh <- true
	}
}

func (s *FusisState) GetServices() []types.Service {
	s.Lock()
	defer s.Unlock()

	services := []types.Service{}
	for _, v := range s.services {
		services = append(services, v)
	}
	return services
}

func (s *FusisState) GetService(name string) (*types.Service, error) {
	svc, ok := s.services[name]
	if !ok {
		return nil, types.ErrServiceNotFound
	}

	return &svc, nil
}

func (s *FusisState) GetDestinations(svc *types.Service) []types.Destination {
	dsts := []types.Destination{}
	for _, d := range s.destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}

	return dsts
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

func (s *FusisState) GetDestination(name string) (*types.Destination, error) {
	s.Lock()
	defer s.Unlock()

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
