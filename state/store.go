package state

import (
	"fmt"
	"time"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/health"
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

	AddCheck(dst *types.Destination)
	DeleteCheck(dst *types.Destination)
	GetChecks() map[string]*health.Check
}

func (s *State) GetServices() []types.Service {
	s.Lock()
	defer s.Unlock()

	services := []types.Service{}
	for _, v := range s.services {
		// s.getDestinations(&v)
		services = append(services, v)
	}
	return services
}

func (s *State) GetService(name string) (*types.Service, error) {
	s.Lock()
	defer s.Unlock()

	svc := s.services[name]
	if svc.Name == "" {
		return nil, types.ErrServiceNotFound
	}
	// s.getDestinations(&svc)
	return &svc, nil
}

func (s *State) GetDestinations(svc *types.Service) []types.Destination {
	dsts := []types.Destination{}
	for _, d := range s.destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}

	return dsts
}

func (s *State) AddService(svc *types.Service) {
	s.Lock()
	defer s.Unlock()

	s.services[svc.GetId()] = *svc
}

func (s *State) DeleteService(svc *types.Service) {
	s.Lock()
	defer s.Unlock()

	delete(s.services, svc.GetId())
}

func (s *State) GetDestination(name string) (*types.Destination, error) {
	s.Lock()
	defer s.Unlock()

	dst := s.destinations[name]
	if dst.Name == "" {
		return nil, types.ErrDestinationNotFound
	}
	return &dst, nil
}

func (s *State) AddDestination(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	s.destinations[dst.GetId()] = *dst
}

func (s *State) DeleteDestination(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	delete(s.destinations, dst.GetId())
}

func (s *State) AddCheck(dst *types.Destination) {
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

func (s *State) DeleteCheck(dst *types.Destination) {
	s.Lock()
	defer s.Unlock()

	s.checks[dst.GetId()].Stop()
	delete(s.checks, dst.GetId())
}

func (s *State) UpdateCheck(check *health.Check) {
	s.Lock()
	defer s.Unlock()

	s.checks[check.DestinationID].Status = check.Status
}

func (s *State) GetChecks() map[string]*health.Check {
	return s.checks
}
