package provider

import (
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
)

type Provider interface {
	AllocateVip(s *ipvs.Service) error
	ReleaseVip(s ipvs.Service) error
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
