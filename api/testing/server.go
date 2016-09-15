package testing

import (
	"net/http/httptest"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/api/types"
)

type testBalancer struct {
	services     []types.Service
	destinations []types.Destination
}

type FakeFusisServer struct {
	*httptest.Server
	Balancer api.Balancer
	api      *api.ApiService
}

func NewFakeFusisServer() *FakeFusisServer {
	balancer := newTestBalancer()
	apiHandler := api.NewAPI(balancer)
	srv := httptest.NewServer(apiHandler)
	return &FakeFusisServer{
		Server:   srv,
		api:      &apiHandler,
		Balancer: balancer,
	}
}

func newTestBalancer() *testBalancer {
	return &testBalancer{}
}

func (b *testBalancer) GetLeader() string {
	return "localhost:8000"
}

func (b *testBalancer) IsLeader() bool {
	return true
}

func (b *testBalancer) GetServices() []types.Service {
	return b.services
}

func (b *testBalancer) AddService(srv *types.Service) error {
	for i := range b.services {
		if b.services[i].Name == srv.Name {
			return types.ErrServiceAlreadyExists
		}
	}
	b.services = append(b.services, *srv)
	return nil
}

func (b *testBalancer) GetService(id string) (*types.Service, error) {
	for i := range b.services {
		if b.services[i].Name == id {
			return &b.services[i], nil
		}
	}
	return nil, types.ErrServiceNotFound
}

func (b *testBalancer) DeleteService(id string) error {
	for i := range b.services {
		if b.services[i].Name == id {
			b.services = append(b.services[:i], b.services[i+1:]...)
			return nil
		}
	}
	return types.ErrServiceNotFound
}

func (b *testBalancer) AddDestination(srv *types.Service, dest *types.Destination) error {
	var foundSrv *types.Service
	for i := range b.services {
		if b.services[i].Name == srv.Name { // Found service
			foundSrv = &b.services[i]
			dsts := b.GetDestinations(&b.services[i])
			for j := range dsts {
				if dsts[j].Name == dest.Name {
					return types.ErrDestinationAlreadyExists
				}
				if dsts[j].Address == dest.Address && dsts[j].Port == dest.Port {
					return types.ErrDestinationAlreadyExists
				}
			}
		}
	}
	if foundSrv == nil {
		return types.ErrServiceNotFound
	}
	b.destinations = append(b.destinations, *dest)
	return nil
}

func (b *testBalancer) GetDestination(id string) (*types.Destination, error) {
	for j := range b.destinations {
		if b.destinations[j].Name == id {
			return &b.destinations[j], nil
		}
	}
	return nil, types.ErrDestinationNotFound
}

func (b *testBalancer) GetDestinations(svc *types.Service) []types.Destination {
	dsts := []types.Destination{}

	for _, d := range b.destinations {
		if d.ServiceId == svc.GetId() {
			dsts = append(dsts, d)
		}
	}

	return dsts
}

func (b *testBalancer) DeleteDestination(dest *types.Destination) error {
	for i := range b.services {
		srv := &b.services[i]
		dsts := b.GetDestinations(srv)
		for j := range dsts {
			if dsts[j].Name == dest.Name {
				b.destinations = append(dsts[:j], dsts[j+1:]...)
				return nil
			}
		}
	}
	return types.ErrDestinationNotFound
}
