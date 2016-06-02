package ipvs

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("Not found")

type State interface {
	GetServices() *[]Service
	GetService(name string) (*Service, error)
	AddService(svc *Service)
	DeleteService(svc *Service)
	//
	// GetDestination(name string) *Destination
	// AddDestination(dst *Destination) error
	// DeleteDestination(dst *Destination) error
}

type FusisState struct {
	sync.Mutex
	Services     map[string]Service
	Destinations map[string]Destination
}

func NewFusisState() *FusisState {
	return &FusisState{
		Services:     make(map[string]Service),
		Destinations: make(map[string]Destination),
	}
}

func (s *FusisState) GetServices() *[]Service {
	services := []Service{}
	for _, v := range s.Services {
		services = append(services, v)
	}

	return &services
}

func (s *FusisState) GetService(name string) (*Service, error) {
	srv := s.Services[name]

	if srv.Name == "" {
		return nil, ErrNotFound
	}

	return &srv, nil
}

func (s *FusisState) AddService(svc *Service) {
	s.Lock()
	defer s.Unlock()

	s.Services[svc.GetId()] = *svc
}

func (s *FusisState) DeleteService(svc *Service) {
	s.Lock()
	defer s.Unlock()

	delete(s.Services, svc.GetId())
}
