package store

import (
	"encoding/json"
	"fmt"

	"github.com/docker/libkv"
	kv "github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/pkg/errors"
)

func init() {
	registryStores()
}

func registryStores() {
	libkv.AddStore(kv.CONSUL, consul.New)
	libkv.AddStore(kv.ETCD, etcd.New)
}

type Store interface {
	// GetServices() ([]types.Service, error)
	// GetService(name string) (*types.Service, error)
	AddService(svc types.Service) error
	DeleteService(svc types.Service) error
	//
	// GetDestination(name string) (*types.Destination, error)
	// GetDestinations(svc *types.Service) []types.Destination
	AddDestination(svc types.Service, dst types.Destination) error
	DeleteDestination(svc types.Service, dst types.Destination) error

	// AddCheck(dst *types.Destination)
	// DeleteCheck(dst *types.Destination)
	// GetChecks() map[string]*health.Check
	WatchServices() error
	GetServicesCh() chan []types.Service

	WatchDestinations() error
	GetDestinationsCh() chan []types.Destination
}

type FusisStore struct {
	kv kv.Store

	ServicesCh     chan []types.Service
	DestinationsCh chan []types.Destination
}

func New(config *config.BalancerConfig) (Store, error) {
	storeAddress := "192.168.151.187:8500"

	kv, err := libkv.NewStore(
		"consul",
		[]string{storeAddress},
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create store consul")
	}

	svcsCh := make(chan []types.Service)
	dstsCh := make(chan []types.Destination)

	return &FusisStore{kv, svcsCh, dstsCh}, nil
}

func (s FusisStore) AddService(svc types.Service) error {
	key := fmt.Sprintf("fusis/services/%s/config", svc.GetId())

	value, err := json.Marshal(svc)
	if err != nil {
		return errors.Wrapf(err, "error marshaling service: %v", svc)
	}

	err = s.kv.Put(key, value, nil)
	if err != nil {
		return errors.Wrapf(err, "error sending service to store: %v", svc)
	}

	return nil
}

func (s FusisStore) DeleteService(svc types.Service) error {
	key := fmt.Sprintf("fusis/services/%s", svc.GetId())

	err := s.kv.DeleteTree(key)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete service: %v", svc)
	}

	return nil
}

func (s FusisStore) GetServicesCh() chan []types.Service {
	return s.ServicesCh
}

func (s FusisStore) WatchServices() error {
	svcs := []types.Service{}

	stopCh := make(<-chan struct{})
	events, err := s.kv.WatchTree("fusis/services", stopCh)
	if err != nil {
		return err
	}

	for {
		select {
		case entries := <-events:
			for _, pair := range entries {
				svc := types.Service{}
				if err := json.Unmarshal(pair.Value, &svc); err != nil {
					return err
				}

				svcs = append(svcs, svc)
			}

			s.ServicesCh <- svcs

			//Cleaning up services slice
			svcs = []types.Service{}
		}
	}

	return nil
}

func (s FusisStore) AddDestination(svc types.Service, dst types.Destination) error {
	key := fmt.Sprintf("fusis/destinations/%s/%s", svc.GetId(), dst.GetId())

	value, err := json.Marshal(dst)
	if err != nil {
		return errors.Wrapf(err, "error marshaling destination: %v", dst)
	}

	err = s.kv.Put(key, value, nil)
	if err != nil {
		return errors.Wrapf(err, "error sending destination to store: %v", dst)
	}

	return nil
}

func (s FusisStore) DeleteDestination(svc types.Service, dst types.Destination) error {
	key := fmt.Sprintf("fusis/destinations/%s/%s", svc.GetId(), dst.GetId())

	err := s.kv.DeleteTree(key)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete destination: %v", dst)
	}

	return nil
}

func (s FusisStore) GetDestinationsCh() chan []types.Destination {
	return s.DestinationsCh
}

func (s FusisStore) WatchDestinations() error {
	dsts := []types.Destination{}

	stopCh := make(<-chan struct{})
	events, err := s.kv.WatchTree("fusis/destinations", stopCh)
	if err != nil {
		return err
	}

	for {
		select {
		case entries := <-events:
			for _, pair := range entries {
				svc := types.Destination{}
				if err := json.Unmarshal(pair.Value, &svc); err != nil {
					return err
				}

				dsts = append(dsts, svc)
			}

			s.DestinationsCh <- dsts

			//Cleaning up destinations slice
			dsts = []types.Destination{}
		}
	}

	return nil
}

// func (s FusisStore) GetServices() ([]types.Service, error) {
// 	svcs := []types.Service{}
//
// 	entries, err := s.kv.List("fusis/services")
// 	for _, pair := range entries {
// 		svc := types.Service{}
// 		if err := json.Unmarshal(pair.Value, &svc); err != nil {
// 			return svcs, err
// 		}
//
// 		svcs = append(svcs, svc)
// 	}
//
// 	return svcs, err
// }
