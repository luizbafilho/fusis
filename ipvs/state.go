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

	GetDestination(name string) (*Destination, error)
	AddDestination(dst *Destination)
	DeleteDestination(dst *Destination)
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
		s.getDestinations(&v)
		services = append(services, v)
	}

	return &services
}

func (s *FusisState) GetService(name string) (*Service, error) {
	svc := s.Services[name]

	if svc.Name == "" {
		return nil, ErrNotFound
	}

	s.getDestinations(&svc)

	return &svc, nil
}

func (s *FusisState) getDestinations(svc *Service) {
	dsts := []Destination{}

	for _, d := range s.Destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}

	svc.Destinations = dsts
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

func (s *FusisState) GetDestination(name string) (*Destination, error) {
	dst := s.Destinations[name]

	if dst.Name == "" {
		return nil, ErrNotFound
	}

	return &dst, nil
}

func (s *FusisState) AddDestination(dst *Destination) {
	s.Lock()
	defer s.Unlock()

	s.Destinations[dst.GetId()] = *dst
}

func (s *FusisState) DeleteDestination(dst *Destination) {
	s.Lock()
	defer s.Unlock()

	delete(s.Destinations, dst.GetId())
}
