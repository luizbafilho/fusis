package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/health"
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

	AddCheck(dst *types.Destination)
	DeleteCheck(dst *types.Destination)
	GetChecks() map[string]*health.Check

	ChangesCh() chan bool
}

// State...
type FusisState struct {
	sync.Mutex

	services     map[string]types.Service
	destinations map[string]types.Destination
	checks       map[string]*health.Check

	changesCh    chan bool
	healhCheckCh chan health.Check

	store store.Store
}

type Services map[string]types.Service
type Destinations map[string]types.Destination

// New creates a new Engine
func New(store store.Store, config *config.BalancerConfig) (State, error) {
	state := &FusisState{
		services:     make(Services),
		destinations: make(Destinations),
		checks:       make(map[string]*health.Check),
		changesCh:    make(chan bool),
		healhCheckCh: make(chan health.Check),
		store:        store,
	}

	state.watchStore()

	return state, nil
}

func (s *FusisState) ChangesCh() chan bool {
	return s.changesCh
}

func (s *FusisState) HealthCheckCh() chan health.Check {
	return s.healhCheckCh
}

func (s *FusisState) watchStore() {
	go s.store.WatchServices()
	go s.store.WatchDestinations()

	go s.handleServicesChange()
	go s.handleDestinationsChange()
}

func (s *FusisState) handleServicesChange() {
	for svcs := range s.store.GetServicesCh() {
		s.services = Services{}
		for _, svc := range svcs {
			s.AddService(svc)
		}

		s.changesCh <- true
	}
}

func (s *FusisState) handleDestinationsChange() {
	for dsts := range s.store.GetDestinationsCh() {
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

func (s *FusisState) AddCheck(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	check := health.Check{
		UpdatesCh:     s.healhCheckCh,
		Status:        health.BAD,
		Interval:      5 * time.Second,
		TCP:           fmt.Sprintf("%s:%d", dst.Address, dst.Port),
		DestinationID: dst.GetId(),
	}

	s.checks[dst.GetId()] = &check

	check.Start()
}

func (s *FusisState) DeleteCheck(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	s.checks[dst.GetId()].Stop()
	delete(s.checks, dst.GetId())
}

func (s *FusisState) UpdateCheck(check *health.Check) {
	s.Lock()
	defer s.Unlock()

	s.checks[check.DestinationID].Status = check.Status
}

func (s *FusisState) GetChecks() map[string]*health.Check {
	return s.checks
}
