package election

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
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

func TestElection(t *testing.T) {
	// cleanup(t)
	// config1 := config.BalancerConfig{
	// 	Name:          "fusis-1",
	// 	EtcdEndpoints: "172.100.0.40:2379",
	// }
	// config2 := config.BalancerConfig{
	// 	Name:          "fusis-2",
	// 	EtcdEndpoints: "172.100.0.40:2379",
	// }
	//
	// el, err := New(&config1)
	// assert.Nil(t, err)
	// el2, err := New(&config2)
	// assert.Nil(t, err)
	//
	// ch1 := make(chan bool)
	// ch2 := make(chan bool)
	//
	// var wg sync.WaitGroup
	// wg.Add(3)
	// go func() {
	// 	defer wg.Done()
	// 	for i := 0; i < 2; i++ {
	// 		select {
	// 		case v1 := <-ch1:
	// 			assert.Equal(t, false, v1)
	// 		case v2 := <-ch2:
	// 			assert.Equal(t, true, v2)
	// 		}
	// 	}
	// }()
	//
	// go func() {
	// 	defer wg.Done()
	// 	time.Sleep(2 * time.Second)
	// 	ch1 = el.Run()
	// }()
	//
	// go func() {
	// 	defer wg.Done()
	// 	ch2 = el2.Run()
	// }()
	//
	// wg.Wait()
	//
	// // Testing resign el2
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for i := 0; i < 2; i++ {
	// 		select {
	// 		case v1 := <-ch1:
	// 			assert.Equal(t, true, v1)
	// 		case v2 := <-ch2:
	// 			assert.Equal(t, false, v2)
	// 		}
	// 	}
	// }()
	//
	// err = el2.Resign()
	// assert.Nil(t, err)
	//
	// wg.Wait()
	//
	// // Testing resign el1
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for i := 0; i < 2; i++ {
	// 		select {
	// 		case v1 := <-ch1:
	// 			assert.Equal(t, false, v1)
	// 		case v2 := <-ch2:
	// 			assert.Equal(t, true, v2)
	// 		}
	// 	}
	// }()
	//
	// err = el.Resign()
	// assert.Nil(t, err)
	//
	// wg.Wait()
}
