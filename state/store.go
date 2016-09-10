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
	GetDestinations(svc *types.Service) []types.Destination
	AddDestination(dst *types.Destination)
	DeleteDestination(dst *types.Destination)
	CollectStats(tick time.Time)
}

type FusisStore struct {
	sync.Mutex
	Services     map[string]types.Service
	Destinations map[string]types.Destination
}

func NewFusisStore() *FusisStore {
	return &FusisStore{
		Services:     make(map[string]types.Service),
		Destinations: make(map[string]types.Destination),
	}
}

func (s *FusisStore) GetServices() []types.Service {
	s.Lock()
	defer s.Unlock()

	services := []types.Service{}
	for _, v := range s.Services {
		// s.getDestinations(&v)
		services = append(services, v)
	}
	return services
}

func (s *FusisStore) GetService(name string) (*types.Service, error) {
	s.Lock()
	defer s.Unlock()

	svc := s.Services[name]
	if svc.Name == "" {
		return nil, types.ErrServiceNotFound
	}
	// s.getDestinations(&svc)
	return &svc, nil
}

func (s *FusisStore) GetDestinations(svc *types.Service) []types.Destination {
	dsts := []types.Destination{}
	for _, d := range s.Destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}

	return dsts
}

func (s *FusisStore) AddService(svc *types.Service) {
	s.Lock()
	defer s.Unlock()

	s.Services[svc.GetId()] = *svc
}

func (s *FusisStore) DeleteService(svc *types.Service) {
	s.Lock()
	defer s.Unlock()

	delete(s.Services, svc.GetId())
}

func (s *FusisStore) GetDestination(name string) (*types.Destination, error) {
	s.Lock()
	defer s.Unlock()

	dst := s.Destinations[name]
	if dst.Name == "" {
		return nil, types.ErrDestinationNotFound
	}
	return &dst, nil
}

func (s *FusisStore) AddDestination(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	s.Destinations[dst.GetId()] = *dst
}

func (s *FusisStore) DeleteDestination(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	delete(s.Destinations, dst.GetId())
}

func (s *FusisStore) CollectStats(tick time.Time) {

}
