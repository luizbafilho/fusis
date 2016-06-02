package provider

import (
	"errors"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
)

type Provider interface {
	AllocateVIP(s *ipvs.Service) error
	ReleaseVIP(s ipvs.Service) error
	AssignVIP(s ipvs.Service) error
	UnassignVIP(s ipvs.Service) error
}

type InitializableProvider interface {
	Initialize(state ipvs.State) error
}

var ErrProviderNotRegistered = errors.New("Provider not registered")

type providerFactory func() Provider

var providerFactories = make(map[string]providerFactory)
var providerInstances = make(map[string]Provider)

func RegisterProviderFactory(ptype string, fac providerFactory) {
	providerFactories[ptype] = fac
}

func New(state ipvs.State) (Provider, error) {
	providerName := config.Balancer.Provider.Type

	instance, ok := providerInstances[providerName]
	if !ok {
		factory, ok := providerFactories[providerName]
		if !ok {
			return nil, ErrProviderNotRegistered
		}

		instance = factory()
		if init, ok := instance.(InitializableProvider); ok {
			if err := init.Initialize(state); err != nil {
				return nil, err
			}
		}

		providerInstances[providerName] = instance
	}

	return instance, nil
}
