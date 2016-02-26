package etcd

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
	"github.com/luizbafilho/janus/store"
)

type Etcd struct {
	client client.KeysAPI
}

func New(addrs []string) store.Store {
	cfg := client.Config{
		Endpoints:               addrs,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	store := Etcd{
		client: client.NewKeysAPI(c),
	}

	return store
}

func (s Etcd) AddService(svc store.ServiceRequest) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s", svc.Host, svc.Port, svc.Protocol)
	value, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	_, err = s.client.Set(context.Background(), key, string(value), nil)

	return err
}
