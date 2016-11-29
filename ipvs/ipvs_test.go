package ipvs_test

import (
	"testing"

	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/types"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/state/mocks"
	"github.com/stretchr/testify/assert"
)

func TestIpvsSync(t *testing.T) {
	svc := types.Service{
		Name:      "",
		Address:   "10.0.1.1",
		Port:      80,
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	dst := types.Destination{
		Name:      "host1",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "service1",
	}

	dst2 := types.Destination{
		Name:      "host2",
		Address:   "192.168.1.2",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "service1",
	}

	state := &mocks.State{}
	state.On("GetServices").Return([]types.Service{svc})
	state.On("GetDestinations", &svc).Return([]types.Destination{dst})

	ipvsMngr, err := ipvs.New()
	assert.Nil(t, err)

	// Testing Services and destinations addition
	ipvsMngr.Sync(state)

	svcs, err := gipvs.GetServices()
	assert.Nil(t, err)

	// Asserting Services
	assert.Equal(t, svc.Address, svcs[0].Address.String())
	assert.Equal(t, svc.Port, svcs[0].Port)

	// Asserting Destinations
	assert.Equal(t, dst.Address, svcs[0].Destinations[0].Address.String())
	assert.Equal(t, dst.Port, svcs[0].Destinations[0].Port)

	// Updating destinations
	state = &mocks.State{}
	state.On("GetServices").Return([]types.Service{svc})
	state.On("GetDestinations", &svc).Return([]types.Destination{dst2})
	ipvsMngr.Sync(state)

	svcs, err = gipvs.GetServices()
	assert.Nil(t, err)

	assert.Equal(t, dst2.Address, svcs[0].Destinations[0].Address.String())
	assert.Equal(t, dst2.Port, svcs[0].Destinations[0].Port)

	// Testing Services and Destinations deletion
	state = &mocks.State{}
	state.On("GetServices").Return([]types.Service{})

	ipvsMngr.Sync(state)

	svcs, err = gipvs.GetServices()
	assert.Nil(t, err)

	assert.Len(t, svcs, 0)

	ipvsMngr.Flush()
}
