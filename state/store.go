package state

import (
	"sync"
	"time"

	"github.com/luizbafilho/fusis/api/types"
)

type Store interface {
	GetServices() []types.Service
	GetService(name string) (*types.Service, error)
	AddService(svc *types.Service)
	DeleteService(svc *types.Service)

	GetDestination(name string) (*types.Destination, error)
	AddDestination(dst *types.Destination)
	DeleteDestination(dst *types.Destination)
	CollectStats(tick time.Time)
}

type FusisState struct {
	sync.Mutex
	Services     map[string]types.Service
	Destinations map[string]types.Destination
}

func NewFusisState() *FusisState {
	return &FusisState{
		Services:     make(map[string]types.Service),
		Destinations: make(map[string]types.Destination),
	}
}

func (s *FusisState) GetServices() []types.Service {
	s.Lock()
	defer s.Unlock()

	services := []types.Service{}
	for _, v := range s.Services {
		s.getDestinations(&v)
		services = append(services, v)
	}
	return services
}

func (s *FusisState) GetService(name string) (*types.Service, error) {
	s.Lock()
	defer s.Unlock()

	svc := s.Services[name]
	if svc.Name == "" {
		return nil, types.ErrServiceNotFound
	}
	s.getDestinations(&svc)
	return &svc, nil
}

func (s *FusisState) getDestinations(svc *types.Service) {
	dsts := []types.Destination{}
	for _, d := range s.Destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}
	svc.Destinations = dsts
}

func (s *FusisState) AddService(svc *types.Service) {
	s.Lock()
	defer s.Unlock()

	s.Services[svc.GetId()] = *svc
}

func (s *FusisState) DeleteService(svc *types.Service) {
	s.Lock()
	defer s.Unlock()

	delete(s.Services, svc.GetId())
}

func (s *FusisState) GetDestination(name string) (*types.Destination, error) {
	s.Lock()
	defer s.Unlock()

	dst := s.Destinations[name]
	if dst.Name == "" {
		return nil, types.ErrDestinationNotFound
	}
	return &dst, nil
}

func (s *FusisState) AddDestination(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	s.Destinations[dst.GetId()] = *dst
}

func (s *FusisState) DeleteDestination(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	delete(s.Destinations, dst.GetId())
}

func (s *FusisState) CollectStats(tick time.Time) {

}
