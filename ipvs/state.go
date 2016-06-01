package ipvs

import "sync"

type State interface {
	GetServices() *[]Service
	// GetService(name string) *Service
	AddService(svc *Service)
	// DeleteService(svc *Service) error
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

func (s *FusisState) AddService(svc *Service) {
	s.Lock()
	defer s.Unlock()

	s.Services[svc.GetId()] = *svc
}
