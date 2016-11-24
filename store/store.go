package store

import (
	"encoding/json"
	"fmt"
	"net/url"

	log "github.com/Sirupsen/logrus"
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
	AddService(svc *types.Service) error
	DeleteService(svc *types.Service) error
	AddDestination(svc *types.Service, dst *types.Destination) error
	DeleteDestination(svc *types.Service, dst *types.Destination) error

	WatchServices()
	GetServicesCh() chan []types.Service

	WatchDestinations() error
	GetDestinationsCh() chan []types.Destination

	GetKV() kv.Store

	// AddCheck(dst *types.Destination)
	// DeleteCheck(dst *types.Destination)
	// GetChecks() map[string]*health.Check
}

var (
	ErrUnsupportedStore = errors.New("[store] Unsupported store")
)

type FusisStore struct {
	kv kv.Store

	ServicesCh     chan []types.Service
	DestinationsCh chan []types.Destination
}

func New(config *config.BalancerConfig) (Store, error) {
	u, err := url.Parse(config.StoreAddress)
	if err != nil {
		return nil, errors.Wrap(err, "error paring store address")
	}

	scheme := u.Scheme
	if scheme != "consul" && scheme != "etcd" {
		return nil, ErrUnsupportedStore
	}

	kv, err := libkv.NewStore(
		kv.Backend(scheme),
		[]string{u.Host},
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create store consul")
	}

	svcsCh := make(chan []types.Service)
	dstsCh := make(chan []types.Destination)

	return &FusisStore{kv, svcsCh, dstsCh}, nil
}

func (s *FusisStore) GetKV() kv.Store {
	return s.kv
}

func (s *FusisStore) AddService(svc *types.Service) error {
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

func (s *FusisStore) DeleteService(svc *types.Service) error {
	key := fmt.Sprintf("fusis/services/%s", svc.GetId())

	err := s.kv.DeleteTree(key)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete service: %v", svc)
	}

	return nil
}

func (s *FusisStore) GetServicesCh() chan []types.Service {
	return s.ServicesCh
}

func (s *FusisStore) WatchServices() {
	svcs := []types.Service{}

	stopCh := make(<-chan struct{})
	events, err := s.kv.WatchTree("fusis/services", stopCh)
	if err != nil {
		log.Error("[store] ", err)
	}

	for {
		select {
		case entries := <-events:
			log.Debug("[store] Services received")

			for _, pair := range entries {
				svc := types.Service{}
				if err := json.Unmarshal(pair.Value, &svc); err != nil {
					log.Error("[store] ", err)
				}
				log.Debugf("[store] Got sevice: %#v ", svc)

				svcs = append(svcs, svc)
			}

			s.ServicesCh <- svcs

			//Cleaning up services slice
			svcs = []types.Service{}
		}
	}
}

func (s *FusisStore) AddDestination(svc *types.Service, dst *types.Destination) error {
	key := fmt.Sprintf("fusis/destinations/%s/%s", svc.GetId(), dst.GetId())

	value, err := json.Marshal(dst)
	if err != nil {
		return errors.Wrapf(err, "error marshaling destination: %v", dst)
	}

	err = s.kv.Put(key, value, nil)
	if err != nil {
		return errors.Wrapf(err, "error sending destination to store: %v", dst)
	}
	log.Debugf("[store] Added destination: %#v with key: %s", value, key)

	return nil
}

func (s *FusisStore) DeleteDestination(svc *types.Service, dst *types.Destination) error {
	key := fmt.Sprintf("fusis/destinations/%s/%s", svc.GetId(), dst.GetId())

	err := s.kv.DeleteTree(key)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete destination: %v", dst)
	}
	log.Debugf("[store] Deleted destination: %s", key)

	return nil
}

func (s *FusisStore) GetDestinationsCh() chan []types.Destination {
	return s.DestinationsCh
}

func (s *FusisStore) WatchDestinations() error {
	dsts := []types.Destination{}

	stopCh := make(<-chan struct{})
	events, err := s.kv.WatchTree("fusis/destinations", stopCh)
	if err != nil {
		return err
	}

	for {
		select {
		case entries := <-events:
			log.Debug("[store] Destinations received")

			for _, pair := range entries {
				dst := types.Destination{}
				if err := json.Unmarshal(pair.Value, &dst); err != nil {
					return err
				}
				log.Debugf("[store] Got destination: %#v", dst)

				dsts = append(dsts, dst)
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
