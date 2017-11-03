package election

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/luizbafilho/fusis/config"
	"github.com/stretchr/testify/assert"
)

var kv *clientv3.Client

func init() {
	var err error
	kv, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"172.100.0.40:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return
		log.Fatal("failed to start etcd connection")
	}
}

func cleanup(t *testing.T) {
	_, err := kv.Delete(context.TODO(), "/test", clientv3.WithPrefix())
	assert.Nil(t, err)
}

func setup(t *testing.T) {
	cleanup(t)
}

func teardown(t *testing.T) {
	err := kv.Close()
	assert.Nil(t, err)
}

func TestElection(t *testing.T) {
	cleanup(t)
	config1 := config.BalancerConfig{
		Name:          "fusis-1",
		EtcdEndpoints: "172.100.0.40:2379",
	}
	config2 := config.BalancerConfig{
		Name:          "fusis-2",
		EtcdEndpoints: "172.100.0.40:2379",
	}

	el, err := New(&config1, "test")
	assert.Nil(t, err)
	el2, err := New(&config2, "test")
	assert.Nil(t, err)

	ch1 := make(chan bool)
	ch2 := make(chan bool)

	go func() {
		time.Sleep(4 * time.Second)
		el.Run(ch1)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		el2.Run(ch2)
	}()

	// el2 elected
	assert.Equal(t, true, <-ch2)

	// el2 resign so el get elected
	err = el2.Resign()
	assert.Nil(t, err)

	assert.Equal(t, true, <-ch1)
	assert.Equal(t, false, <-ch2)

	// el resign so el2 get elected
	err = el.Resign()
	assert.Nil(t, err)

	assert.Equal(t, false, <-ch1)
	assert.Equal(t, true, <-ch2)
}
