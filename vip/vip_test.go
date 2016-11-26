package vip_test

import (
	"testing"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state/mocks"
	"github.com/luizbafilho/fusis/vip"
	"github.com/stretchr/testify/assert"
)

func TestVipsSync(t *testing.T) {
	s1 := types.Service{
		Name:     "test",
		Address:  "10.0.1.1",
		Port:     80,
		Mode:     "nat",
		Protocol: "tcp",
	}

	s2 := types.Service{
		Name:     "test2",
		Address:  "10.0.1.2",
		Port:     80,
		Protocol: "tcp",
		Mode:     "nat",
	}

	state := &mocks.State{}
	state.On("GetServices").Return([]types.Service{s1, s2})

	iface := "eth0"
	config := &config.BalancerConfig{
		Interfaces: config.Interfaces{
			Inbound:  iface,
			Outbound: iface,
		},
	}
	vipMngr, err := vip.New(config)
	assert.Nil(t, err)

	vipMngr.Sync(state)

	vips, err := net.GetFusisVipsIps(iface)
	assert.Nil(t, err)

	assert.Len(t, vips, 2)
	assert.Contains(t, vips, "10.0.1.1")
	assert.Contains(t, vips, "10.0.1.2")

	// Asserting remove
	state = &mocks.State{}
	state.On("GetServices").Return([]types.Service{s2})

	err = vipMngr.Sync(state)

	vips, err = net.GetFusisVipsIps(iface)
	assert.Nil(t, err)
	assert.Len(t, vips, 1)
	assert.Contains(t, vips, "10.0.1.2")

	net.DelVips(iface)
}
