package store

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/types"
	"github.com/stretchr/testify/assert"
)

func cleanup(t *testing.T) {
	kv, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"172.100.0.40:2379"},
		DialTimeout: 5 * time.Second,
	})
	assert.Nil(t, err)

	_, err = kv.Delete(context.TODO(), "/", clientv3.WithPrefix())
	assert.Nil(t, err)
}

func TestPutAndGetServices(t *testing.T) {
	cleanup(t)
	config := config.BalancerConfig{
		EtcdEndpoints: "172.100.0.40:2379",
	}

	kv, err := New(&config)
	assert.Nil(t, err)

	svc1 := types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	svc2 := types.Service{
		Name:      "test2",
		Address:   "10.0.1.2",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	err = kv.AddService(&svc1)
	assert.Nil(t, err)

	err = kv.AddService(&svc2)
	assert.Nil(t, err)

	// Testing GetServices()
	svcs, err := kv.GetServices()
	assert.Nil(t, err)
	assert.Equal(t, []types.Service{svc1, svc2}, svcs)

	// Testing GetService(name int)
	svc, err := kv.GetService(svc1.Name)
	assert.Nil(t, err)
	assert.Equal(t, &svc1, svc)

	svc, err = kv.GetService("foo")
	assert.Equal(t, types.ErrServiceNotFound, err)

	err = kv.AddService(&svc2)
	assert.NotNil(t, err)
	assert.Equal(t, types.ErrValidation{Type: "service", Errors: map[string]string{"ipvs": "service must be unique"}}, err)

	err = kv.Close()
	assert.Nil(t, err)
}

func TestDeleteService(t *testing.T) {
	cleanup(t)
	config := config.BalancerConfig{
		EtcdEndpoints: "172.100.0.40:2379",
	}

	kv, err := New(&config)
	assert.Nil(t, err)

	svc1 := &types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	svc2 := &types.Service{
		Name:      "test2",
		Address:   "10.0.1.2",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	dst1 := &types.Destination{
		Name:      "dst1",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}

	err = kv.AddService(svc1)
	assert.Nil(t, err)

	err = kv.AddService(svc2)
	assert.Nil(t, err)

	err = kv.AddDestination(svc2, dst1)
	assert.Nil(t, err)

	err = kv.DeleteService(svc1)
	assert.Nil(t, err)

	err = kv.DeleteService(svc2)
	assert.Nil(t, err)

	svcs, err := kv.GetServices()
	assert.Nil(t, err)

	assert.Equal(t, []types.Service{}, svcs)

	dsts, err := kv.GetDestinations(svc2)
	assert.Nil(t, err)
	assert.Equal(t, []types.Destination{}, dsts)

	err = kv.DeleteService(&types.Service{Name: "foo"})
	assert.Equal(t, types.ErrServiceNotFound, err)

	err = kv.Close()
	assert.Nil(t, err)
}

func TestPutAndGetDestinations(t *testing.T) {
	cleanup(t)
	config := config.BalancerConfig{
		EtcdEndpoints: "172.100.0.40:2379",
	}

	kv, err := New(&config)
	assert.Nil(t, err)

	svc1 := &types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	svc2 := &types.Service{
		Name:      "test2",
		Address:   "10.0.1.2",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	err = kv.AddService(svc1)
	assert.Nil(t, err)
	err = kv.AddService(svc2)
	assert.Nil(t, err)

	dst1 := &types.Destination{
		Name:      "dst1",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}

	dst2 := &types.Destination{
		Name:      "dst2",
		Address:   "192.168.1.2",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}

	dst3 := &types.Destination{
		Name:      "dst3",
		Address:   "192.168.1.3",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test2",
	}

	err = kv.AddDestination(svc1, dst1)
	assert.Nil(t, err)
	err = kv.AddDestination(svc1, dst2)
	assert.Nil(t, err)
	err = kv.AddDestination(svc2, dst3)
	assert.Nil(t, err)

	dsts, err := kv.GetDestinations(svc1)
	assert.Nil(t, err)
	assert.Equal(t, []types.Destination{*dst1, *dst2}, dsts)

	dsts, err = kv.GetDestinations(svc2)
	assert.Nil(t, err)
	assert.Equal(t, []types.Destination{*dst3}, dsts)

	err = kv.Close()
	assert.Nil(t, err)
}

func TestDeleteDestination(t *testing.T) {
	cleanup(t)
	config := config.BalancerConfig{
		EtcdEndpoints: "172.100.0.40:2379",
	}

	kv, err := New(&config)
	assert.Nil(t, err)

	svc1 := &types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	dst1 := &types.Destination{
		Name:      "dst1",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}

	err = kv.AddService(svc1)
	assert.Nil(t, err)

	err = kv.AddDestination(svc1, dst1)
	assert.Nil(t, err)

	err = kv.DeleteDestination(svc1, dst1)
	assert.Nil(t, err)

	dsts, err := kv.GetDestinations(svc1)
	assert.Nil(t, err)
	assert.Equal(t, []types.Destination{}, dsts)

	err = kv.DeleteDestination(svc1, &types.Destination{Name: "foo"})
	assert.Equal(t, types.ErrDestinationNotFound, err)

	err = kv.Close()
	assert.Nil(t, err)
}

func TestGetState(t *testing.T) {
	cleanup(t)
	config := config.BalancerConfig{
		EtcdEndpoints: "172.100.0.40:2379",
	}

	kv, err := New(&config)
	assert.Nil(t, err)

	svc1 := &types.Service{
		Name:      "test1",
		Address:   "10.0.1.1",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	dst1 := &types.Destination{
		Name:      "dst1",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}

	err = kv.AddService(svc1)
	assert.Nil(t, err)

	err = kv.AddDestination(svc1, dst1)
	assert.Nil(t, err)

	svc2 := &types.Service{
		Name:      "srv2",
		Address:   "10.0.1.2",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	err = kv.AddService(svc2)
	assert.Nil(t, err)

	fstate, err := kv.GetState()
	assert.Nil(t, err)

	assert.Equal(t, []types.Service{*svc1, *svc2}, fstate.GetServices())
	assert.Equal(t, []types.Destination{*dst1}, fstate.GetDestinations(svc1))

	err = kv.Close()
	assert.Nil(t, err)
}

func TestWatch(t *testing.T) {
	cleanup(t)
	config := config.BalancerConfig{
		EtcdEndpoints: "172.100.0.40:2379",
	}

	kv, err := New(&config)
	assert.Nil(t, err)

	svc1 := &types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	dst1 := &types.Destination{
		Name:      "dst1",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}

	err = kv.AddService(svc1)
	assert.Nil(t, err)

	err = kv.AddDestination(svc1, dst1)
	assert.Nil(t, err)

	sts, err := kv.GetState()
	assert.Nil(t, err)

	ch := make(chan state.State)

	kv.AddWatcher(ch)
	go kv.Watch()

	// wait for the watch
	time.Sleep(4 * time.Second)

	svc2 := &types.Service{
		Name:      "srv2",
		Address:   "10.0.1.2",
		Port:      80,
		Mode:      "nat",
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	err = kv.AddService(svc2)
	assert.Nil(t, err)

	newState := <-ch

	assert.Equal(t, []types.Service{*svc1, *svc2}, newState.GetServices())

	err = kv.Close()
	assert.Nil(t, err)
}
