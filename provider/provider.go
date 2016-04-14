package provider

import "github.com/luizbafilho/fusis/config"

type Provider interface {
	SetVip() (interface{}, error)
	UnsetVip(setResult interface{}) error
}

type providerFactory func() Provider

var providerFactories = make(map[string]providerFactory)

func RegisterProviderFactory(ptype string, fac providerFactory) {
	providerFactories[ptype] = fac
}

func GetProvider() Provider {
	factory := providerFactories[config.BalancerConf.Provider.Type]
	provider := factory()
	return provider
}
