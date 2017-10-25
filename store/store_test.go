package store

import (
	"testing"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/types"
	"github.com/stretchr/testify/assert"
)

func TestGetServices(t *testing.T) {
	config := config.BalancerConfig{
		EtcdEndpoints: "http://172.100.0.40:2379",
	}

	kv, err := New(&config)
	assert.Nil(t, err)

	// svc := types.Service{
	// 	Name:      "test",
	// 	Address:   "10.0.1.1",
	// 	Port:      80,
	// 	Mode:      "nat",
	// 	Scheduler: "lc",
	// 	Protocol:  "tcp",
	// }

	// err = kv.AddService(&svc)
	// assert.Nil(t, err)

	svcs, err := kv.GetServices()
	assert.Nil(t, err)

	assert.Equal(t, []types.Service{}, svcs)
}
