package provider

import (
	"errors"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
)

var ErrProviderNotRegistered = errors.New("Provider not registered")

type Provider interface {
	AllocateVIP(s *ipvs.Service, state ipvs.State) error
	ReleaseVIP(s ipvs.Service) error
	AssignVIP(s *ipvs.Service) error
	UnassignVIP(s *ipvs.Service) error
}

func New(config *config.BalancerConfig) (Provider, error) {
	var provider Provider
	var err error

	switch config.Provider.Type {
	case "none":
		provider, err = NewNone(config)
	}

	return provider, err
}
