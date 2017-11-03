package fusis

import (
	"time"

	"github.com/luizbafilho/fusis/types"
)

// GetServices get all services
func (b *FusisBalancer) GetServices() ([]types.Service, error) {
	return b.store.GetServices()
}

//GetService get a service
func (b *FusisBalancer) GetService(name string) (*types.Service, error) {
	return b.store.GetService(name)
}

// AddService ...
func (b *FusisBalancer) AddService(svc *types.Service) error {
	// Allocate a new VIP if no address provided
	if svc.Address == "" {
		if err := b.ipam.AllocateVIP(svc); err != nil {
			return err
		}
	}

	return b.store.AddService(svc)
}

func (b *FusisBalancer) DeleteService(name string) error {
	svc, err := b.store.GetService(name)
	if err != nil {
		return err
	}

	return b.store.DeleteService(svc)
}

func (b *FusisBalancer) GetDestinations(svc *types.Service) ([]types.Destination, error) {
	return b.store.GetDestinations(svc)
}

func (b *FusisBalancer) AddDestination(svc *types.Service, dst *types.Destination) error {
	// Set defaults
	if dst.Weight == 0 {
		dst.Weight = 1
	}
	if dst.Mode == "" {
		dst.Mode = "nat"
	}

	return b.store.AddDestination(svc, dst)
}

func (b *FusisBalancer) DeleteDestination(dst *types.Destination) error {
	svc, err := b.store.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	return b.store.DeleteDestination(svc, dst)
}

func (b *FusisBalancer) AddCheck(check types.CheckSpec) error {
	// Setting default values
	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}

	if check.Interval == 0 {
		check.Interval = 10 * time.Second
	}

	return b.store.AddCheck(check)
}

func (b *FusisBalancer) DeleteCheck(check types.CheckSpec) error {
	return b.store.DeleteCheck(check)
}
