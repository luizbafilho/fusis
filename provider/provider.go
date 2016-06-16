package provider

import (
	"errors"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
)

var ErrProviderNotRegistered = errors.New("Provider not registered")

type Provider interface {
	AllocateVIP(s *types.Service, state ipvs.State) error
	ReleaseVIP(s types.Service) error
	AssignVIP(s *types.Service) error
	UnassignVIP(s *types.Service) error
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
