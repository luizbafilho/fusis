package provider

import (
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/engine/store"
)

type Provider interface {
	AllocateVip(s *store.Service) error
	ReleaseVip(s store.Service) error
}

type providerFactory func() Provider

var providerFactories = make(map[string]providerFactory)

func RegisterProviderFactory(ptype string, fac providerFactory) {
	providerFactories[ptype] = fac
}

func GetProvider() Provider {
	factory := providerFactories[config.Balancer.Provider.Type]
	provider := factory()
	return provider
}
