package ipam_test

import (
	"testing"

	"github.com/luizbafilho/fusis/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipam"
	"github.com/luizbafilho/fusis/state/mocks"
	"github.com/stretchr/testify/assert"
)

func TestIpAllocationFirstEmpty(t *testing.T) {
	service := types.Service{
		Name:    "test",
		Address: "10.0.1.2",
	}

	state := &mocks.State{}
	state.On("GetServices").Return([]types.Service{service})

	config := &config.BalancerConfig{
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/30", "10.0.1.0/30"},
		},
	}
	ipMngr, err := ipam.New(state, config)
	assert.Nil(t, err)

	err = ipMngr.AllocateVIP(&service)
	assert.Nil(t, err)
	assert.Equal(t, "10.0.1.1", service.Address)
}

func TestIpAllocation(t *testing.T) {
	service1 := types.Service{
		Name:    "test",
		Address: "10.0.1.1",
	}

	service2 := types.Service{
		Name:    "test",
		Address: "10.0.1.2",
	}

	state := &mocks.State{}
	state.On("GetServices").Return([]types.Service{service1, service2})

	config := &config.BalancerConfig{
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/30", "10.0.1.0/30"},
		},
	}
	ipMngr, err := ipam.New(state, config)
	assert.Nil(t, err)

	service3 := types.Service{
		Name: "teste",
	}
	err = ipMngr.AllocateVIP(&service3)
	assert.Nil(t, err)
	assert.Equal(t, "10.0.1.3", service3.Address)
}

func TestIpAllocationMultiplesRanges(t *testing.T) {
	svc1 := types.Service{Address: "10.0.1.1"}
	svc2 := types.Service{Address: "10.0.1.2"}
	svc3 := types.Service{Address: "10.0.1.3"}

	state := &mocks.State{}
	state.On("GetServices").Return([]types.Service{svc1, svc2, svc3})

	config := &config.BalancerConfig{
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/30", "10.0.1.0/30"},
		},
	}
	ipMngr, err := ipam.New(state, config)
	assert.Nil(t, err)

	service := &types.Service{
		Name: "test10",
	}
	err = ipMngr.AllocateVIP(service)
	assert.Nil(t, err)
	assert.Equal(t, "192.168.0.1", service.Address)
}

func TestCursorValidation(t *testing.T) {
	config := &config.BalancerConfig{}

	state := &mocks.State{}
	i, err := ipam.New(state, config)
	assert.Nil(t, err)

	service := &types.Service{
		Name: "test",
	}

	err = i.AllocateVIP(service)
	assert.Equal(t, ipam.ErrNoVipAvailable, err)
}
