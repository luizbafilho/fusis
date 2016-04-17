package provider

import (
	"errors"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
)

type Provider interface {
	AllocateVip(s *ipvs.Service) error
	ReleaseVip(s ipvs.Service) error
}

type InitializableProvider interface {
	Initialize() error
}

var ErrProviderNotRegistered = errors.New("Provider not registered")

type providerFactory func() Provider

var providerFactories = make(map[string]providerFactory)
var providerInstances = make(map[string]Provider)

func RegisterProviderFactory(ptype string, fac providerFactory) {
	providerFactories[ptype] = fac
}

func GetProvider() (Provider, error) {
	providerName := config.Balancer.Provider.Type

	instance, ok := providerInstances[providerName]
	if !ok {
		factory, ok := providerFactories[providerName]
		if !ok {
			return nil, ErrProviderNotRegistered
		}

		instance = factory()
		if init, ok := instance.(InitializableProvider); ok {
			if err := init.Initialize(); err != nil {
				return nil, err
			}
		}

		providerInstances[providerName] = instance
	}

	return instance, nil
}
