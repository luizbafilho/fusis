package testing

import (
	"net/http/httptest"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/api/types"
)

type testBalancer struct {
	services []types.Service
	ids      map[string]int
	dests    map[string][]int
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
	return &testBalancer{
		ids:   make(map[string]int),
		dests: make(map[string][]int),
	}
}

func (b *testBalancer) GetServices() []types.Service {
	return b.services
}

func (b *testBalancer) AddService(srv *types.Service) error {
	_, exists := b.ids[srv.Name]
	if exists {
		return types.ErrServiceAlreadyExists
	}
	b.services = append(b.services, *srv)
	b.ids[srv.Name] = len(b.services) - 1
	return nil
}

func (b *testBalancer) GetService(id string) (*types.Service, error) {
	idx, exists := b.ids[id]
	if !exists {
		return nil, types.ErrServiceNotFound
	}
	return &b.services[idx], nil
}

func (b *testBalancer) DeleteService(id string) error {
	idx, exists := b.ids[id]
	if !exists {
		return types.ErrServiceNotFound
	}
	delete(b.ids, id)
	b.services = append(b.services[:idx], b.services[idx+1:]...)
	return nil
}

func (b *testBalancer) AddDestination(srv *types.Service, dest *types.Destination) error {
	idx, exists := b.ids[srv.Name]
	if !exists {
		return types.ErrServiceNotFound
	}
	_, exists = b.dests[dest.Name]
	if exists {
		return types.ErrDestinationAlreadyExists
	}
	srv = &b.services[idx]
	srv.Destinations = append(srv.Destinations, *dest)
	b.dests[dest.Name] = []int{idx, len(srv.Destinations) - 1}
	return nil
}

func (b *testBalancer) GetDestination(id string) (*types.Destination, error) {
	indexes, exists := b.dests[id]
	if !exists {
		return nil, types.ErrDestinationNotFound
	}
	return &b.services[indexes[0]].Destinations[indexes[1]], nil
}

func (b *testBalancer) DeleteDestination(dest *types.Destination) error {
	indexes, exists := b.dests[dest.Name]
	if !exists {
		return types.ErrDestinationNotFound
	}
	srv := &b.services[indexes[0]]
	srv.Destinations = append(srv.Destinations[:indexes[1]], srv.Destinations[indexes[1]:]...)
	delete(b.dests, dest.Name)
	return nil
}
