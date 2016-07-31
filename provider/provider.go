package provider

import (
	"errors"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/state"
)

var ErrProviderNotRegistered = errors.New("Provider not registered")

type Syncer interface {
	Sync(state state.State) error
}

type Provider interface {
	AllocateVIP(s *types.Service, state state.Store) error
	ReleaseVIP(s types.Service) error
	Syncer
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
