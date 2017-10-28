package store

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"
)

type DistributedLocker struct {
	sessionID string
	ttl       int
}

type Store interface {
	GetState() (state.State, error)
	GetServices() ([]types.Service, error)
	AddService(svc *types.Service) error
	DeleteService(svc *types.Service) error
	// SubscribeServices(ch chan []types.Service)
	// WatchServices()
	//
	GetDestinations(svc *types.Service) ([]types.Destination, error)
	AddDestination(svc *types.Service, dst *types.Destination) error
	DeleteDestination(svc *types.Service, dst *types.Destination) error

	AddWatcher(ch chan state.State)
	Watch()
	// SubscribeDestinations(ch chan []types.Destination)
	// WatchDestinations()
	//
	// AddCheck(check types.CheckSpec) error
	// DeleteCheck(check types.CheckSpec) error
	// SubscribeChecks(ch chan []types.CheckSpec)
	// WatchChecks()
}

type FusisStore struct {
	sync.Mutex
	kv *clientv3.Client

	prefix string

	validate            *validator.Validate
	servicesChannels    []chan []types.Service
	destinationChannels []chan []types.Destination
	checksChannels      []chan []types.CheckSpec
	watchChannels       []chan state.State
}

func New(config *config.BalancerConfig) (Store, error) {
	kv, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(config.EtcdEndpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "connection to etcd failed")
	}

	svcsChs := []chan []types.Service{}
	dstsChs := []chan []types.Destination{}
	checksChs := []chan []types.CheckSpec{}
	watchChs := []chan state.State{}

	validate := validator.New()
	// Registering custom validations
	validate.RegisterValidation("protocols", validateValues(types.Protocols))
	validate.RegisterValidation("schedulers", validateValues(types.Schedulers))

	fusisStore := &FusisStore{
		kv:                  kv,
		prefix:              config.StorePrefix,
		validate:            validate,
		servicesChannels:    svcsChs,
		destinationChannels: dstsChs,
		checksChannels:      checksChs,
		watchChannels:       watchChs,
	}

	// go fusisStore.WatchServices()
	// go fusisStore.WatchDestinations()
	// go fusisStore.WatchChecks()

	return fusisStore, nil
}

func (s *FusisStore) AddWatcher(ch chan state.State) {
	s.Lock()
	s.watchChannels = append(s.watchChannels, ch)
	s.Unlock()
}

func (s *FusisStore) Watch() {
	ticker := time.Tick(1 * time.Second)
	notify := false

	rch := s.kv.Watch(context.TODO(), s.key(""), clientv3.WithPrefix())
	for {
		select {
		case <-rch:
			notify = true
		case <-ticker:
			if !notify {
				continue
			}

			notify = false
			fstate, err := s.GetState()
			if err != nil {
				logrus.Error(err)
			}

			for _, ch := range s.watchChannels {
				ch <- fstate
			}
		}
	}
}

func (s *FusisStore) GetState() (state.State, error) {
	fstate, _ := state.New()

	resp, err := s.kv.Get(context.TODO(), s.key(""), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "[store] Get data from etcd failed")
	}

	for _, pair := range resp.Kvs {

		key := string(pair.Key)
		switch {
		case strings.Contains(key, "ipvs-ids"):
			continue
		case strings.Contains(key, s.key("services")):
			svc := types.Service{}
			if err := json.Unmarshal(pair.Value, &svc); err != nil {
				return nil, errors.Wrapf(err, "[store] Unmarshal %s failed", pair.Value)
			}
			fstate.AddService(svc)
		case strings.Contains(key, s.key("destinations")):
			dst := types.Destination{}
			if err := json.Unmarshal(pair.Value, &dst); err != nil {
				return nil, errors.Wrapf(err, "[store] Unmarshal %s failed", pair.Value)
			}
			fstate.AddDestination(dst)
		}
	}

	return fstate, nil
}

func (s *FusisStore) GetServices() ([]types.Service, error) {
	svcs := []types.Service{}
	resp, err := s.kv.Get(context.TODO(), s.key("services"), clientv3.WithPrefix())
	if err != nil {
		return svcs, errors.Wrap(err, "[store] Get services failed")
	}

	for _, pair := range resp.Kvs {
		svc := types.Service{}
		if err := json.Unmarshal(pair.Value, &svc); err != nil {
			return svcs, errors.Wrapf(err, "[store] Unmarshal %s failed", pair.Value)
		}

		svcs = append(svcs, svc)
	}

	return svcs, nil
}

func (s *FusisStore) GetDestinations(svc *types.Service) ([]types.Destination, error) {
	dsts := []types.Destination{}
	resp, err := s.kv.Get(context.TODO(), s.key("destinations", svc.GetId(), ""), clientv3.WithPrefix())
	if err != nil {
		return dsts, errors.Wrap(err, "[store] Get destinations failed")
	}

	for _, pair := range resp.Kvs {
		dst := types.Destination{}
		if err := json.Unmarshal(pair.Value, &dst); err != nil {
			return dsts, errors.Wrapf(err, "[store] Unmarshal %s failed", pair.Value)
		}

		dsts = append(dsts, dst)
	}

	return dsts, nil
}

//
// AddService adds a new services to store. It validates the name
// uniqueness and the IPVS uniqueness by saving the IPVS key
// in the store, which consists in a combination of address, port
// and protocol.
func (s *FusisStore) AddService(svc *types.Service) error {
	svcKey := s.key("services", svc.GetId(), "config")
	ipvsKey := s.key("ipvs-ids", "services", svc.IpvsId())
	// Validating service
	if err := s.validateService(svc); err != nil {
		return err
	}

	value, err := json.Marshal(svc)
	if err != nil {
		return errors.Wrapf(err, "[store] Error marshaling service: %v", svc)
	}

	svcCmp := clientv3.Compare(clientv3.Version(svcKey), "=", 0)
	ipvsCmp := clientv3.Compare(clientv3.Version(ipvsKey), "=", 0)

	svcPut := clientv3.OpPut(svcKey, string(value))
	ipvsPut := clientv3.OpPut(ipvsKey, "true")

	resp, err := s.kv.Txn(context.TODO()).
		If(svcCmp, ipvsCmp).
		Then(svcPut, ipvsPut).
		Commit()
	if err != nil {
		return errors.Wrapf(err, "[store] Error sending service to Etcd: %v", svc)
	}

	// resp.Succeeded means the compare clause was true
	if resp.Succeeded == false {
		return errors.Errorf("[store] Service or IPVS Service ID are not unique")
	}

	return nil
}

//
// func (s *FusisStore) GetService(serviceId string) (*types.Service, error) {
// 	key := s.key("services", serviceId, "config")
// 	kvPair, err := s.kv.Get(key)
// 	if err != nil {
// 		if err == kv.ErrKeyNotFound {
// 			return nil, types.ErrServiceNotFound
// 		}
// 		return nil, errors.Wrap(err, "GetService from store failed")
// 	}
//
// 	svc := &types.Service{}
// 	if err := json.Unmarshal(kvPair.Value, svc); err != nil {
// 		return nil, errors.Wrap(err, "Unmarshal service from store failed")
// 	}
//
// 	return svc, nil
// }
//
func (s *FusisStore) DeleteService(svc *types.Service) error {
	// Deleting service
	svcKey := s.key("services", svc.GetId())
	ipvsSvcKey := s.key("ipvs-ids", "services", svc.IpvsId())
	svcDel := clientv3.OpDelete(svcKey, clientv3.WithPrefix())
	ipvsSvcDel := clientv3.OpDelete(ipvsSvcKey, clientv3.WithPrefix())

	// Deleting destinations
	dstKey := s.key("destinations", svc.GetId(), "")
	ipvsDstKey := s.key("ipvs-ids", "destinations", svc.GetId(), "")
	dstDel := clientv3.OpDelete(dstKey, clientv3.WithPrefix())
	ipvsDstDel := clientv3.OpDelete(ipvsDstKey, clientv3.WithPrefix())

	_, err := s.kv.Txn(context.TODO()).
		Then(svcDel, ipvsSvcDel, dstDel, ipvsDstDel).
		Commit()
	if err != nil {
		return errors.Wrapf(err, "[store] Error deleting service from Etcd: %v", svc)
	}

	return nil
}

//
// func (s *FusisStore) SubscribeServices(updateCh chan []types.Service) {
// 	s.Lock()
// 	defer s.Unlock()
// 	s.servicesChannels = append(s.servicesChannels, updateCh)
// }
//
// func (s *FusisStore) WatchServices() {
// 	svcs := []types.Service{}
//
// 	stopCh := make(<-chan struct{})
// 	events, err := s.kv.WatchTree(s.key("services"), stopCh)
// 	if err != nil {
// 		log.Error("[store] ", err)
// 	}
//
// 	for {
// 		select {
// 		case entries := <-events:
// 			log.Debug("[store] Services received")
//
// 			for _, pair := range entries {
// 				svc := types.Service{}
// 				if err := json.Unmarshal(pair.Value, &svc); err != nil {
// 					log.Error("[store] ", err)
// 				}
// 				log.Debugf("[store] Got sevice: %#v ", svc)
//
// 				svcs = append(svcs, svc)
// 			}
//
// 			s.Lock()
// 			for _, ch := range s.servicesChannels {
// 				log.Debug("[store] Sending message to service ch")
// 				ch <- svcs
// 			}
// 			s.Unlock()
//
// 			//Cleaning up services slice
// 			svcs = []types.Service{}
// 		}
// 	}
// }
//
func (s *FusisStore) AddDestination(svc *types.Service, dst *types.Destination) error {
	dstKey := s.key("destinations", svc.GetId(), dst.GetId())
	ipvsKey := s.key("ipvs-ids", "destinations", svc.GetId(), dst.IpvsId())
	// Validating destination
	if err := s.validateDestination(dst); err != nil {
		return err
	}

	// Persisting destination
	value, err := json.Marshal(dst)
	if err != nil {
		return errors.Wrapf(err, "[store] error marshaling destination: %v", dst)
	}

	dstCmp := clientv3.Compare(clientv3.Version(dstKey), "=", 0)
	ipvsCmp := clientv3.Compare(clientv3.Version(ipvsKey), "=", 0)

	dstPut := clientv3.OpPut(dstKey, string(value))
	ipvsPut := clientv3.OpPut(ipvsKey, "true")

	resp, err := s.kv.Txn(context.TODO()).
		If(dstCmp, ipvsCmp).
		Then(dstPut, ipvsPut).
		Commit()
	if err != nil {
		return errors.Wrapf(err, "[store] Error sending destination to Etcd: %v", svc)
	}

	// resp.Succeeded means the compare clause was true
	if resp.Succeeded == false {
		return errors.Errorf("[store] Destination or IPVS Destionation ID are not unique")
	}

	return nil
}

func (s *FusisStore) DeleteDestination(svc *types.Service, dst *types.Destination) error {
	dstKey := s.key("destinations", svc.GetId(), dst.GetId())
	ipvsKey := s.key("ipvs-ids", "destinations", svc.GetId(), dst.IpvsId())

	dstDel := clientv3.OpDelete(dstKey, clientv3.WithPrefix())
	ipvsDel := clientv3.OpDelete(ipvsKey, clientv3.WithPrefix())

	_, err := s.kv.Txn(context.TODO()).
		Then(dstDel, ipvsDel).
		Commit()
	if err != nil {
		return errors.Wrapf(err, "[store] Error deleting destination from Etcd: %v", svc)
	}

	return nil
}

//
// func (s *FusisStore) SubscribeDestinations(updateCh chan []types.Destination) {
// 	s.Lock()
// 	defer s.Unlock()
// 	s.destinationChannels = append(s.destinationChannels, updateCh)
// }
//
// func (s *FusisStore) WatchDestinations() {
// 	dsts := []types.Destination{}
//
// 	stopCh := make(<-chan struct{})
// 	events, err := s.kv.WatchTree(s.key("destinations"), stopCh)
// 	if err != nil {
// 		errors.Wrap(err, "failed watching fusis/destinations")
// 	}
//
// 	for {
// 		select {
// 		case entries := <-events:
// 			log.Debug("[store] Destinations received")
//
// 			for _, pair := range entries {
// 				dst := types.Destination{}
// 				if err := json.Unmarshal(pair.Value, &dst); err != nil {
// 					errors.Wrap(err, "failed unmarshall of destinations")
// 				}
// 				log.Debugf("[store] Got destination: %#v", dst)
//
// 				dsts = append(dsts, dst)
// 			}
//
// 			s.Lock()
// 			for _, ch := range s.destinationChannels {
// 				ch <- dsts
// 			}
// 			s.Unlock()
//
// 			//Cleaning up destinations slice
// 			dsts = []types.Destination{}
// 		}
// 	}
// }
//
// func (s *FusisStore) SubscribeChecks(updateCh chan []types.CheckSpec) {
// 	s.Lock()
// 	defer s.Unlock()
// 	s.checksChannels = append(s.checksChannels, updateCh)
// }
//
// func (s *FusisStore) AddCheck(spec types.CheckSpec) error {
// 	key := s.key("checks", spec.ServiceID)
//
// 	value, err := json.Marshal(spec)
// 	if err != nil {
// 		return errors.Wrapf(err, "error marshaling CheckSpec: %#v", spec)
// 	}
//
// 	err = s.kv.Put(key, value, nil)
// 	if err != nil {
// 		return errors.Wrapf(err, "error sending CheckSpec to store: %#v", spec)
// 	}
//
// 	return nil
// }
//
// func (s *FusisStore) DeleteCheck(spec types.CheckSpec) error {
// 	key := s.key("checks", spec.ServiceID)
//
// 	err := s.kv.DeleteTree(key)
// 	if err != nil {
// 		return errors.Wrapf(err, "error trying to delete check: %#v", spec)
// 	}
// 	log.Debugf("[store] Deleted check: %s", key)
//
// 	return nil
// }
//
// func (s *FusisStore) WatchChecks() {
// 	specs := []types.CheckSpec{}
//
// 	stopCh := make(<-chan struct{})
// 	events, err := s.kv.WatchTree(s.key("checks"), stopCh)
// 	if err != nil {
// 		log.Error(errors.Wrap(err, "failed watching fusis/checks"))
// 	}
//
// 	for {
// 		select {
// 		case entries := <-events:
// 			log.Debug("[store] Checks received")
//
// 			for _, pair := range entries {
// 				spec := types.CheckSpec{}
// 				if err := json.Unmarshal(pair.Value, &spec); err != nil {
// 					log.Error(errors.Wrap(err, "failed unmarshall of checks"))
// 				}
// 				log.Debugf("[store] Got Check: %#v", spec)
//
// 				specs = append(specs, spec)
// 			}
//
// 			s.Lock()
// 			for _, ch := range s.checksChannels {
// 				ch <- specs
// 			}
// 			s.Unlock()
//
// 			//Cleaning up destinations slice
// 			specs = []types.CheckSpec{}
// 		}
// 	}
// }
//
func (s *FusisStore) key(keys ...string) string {
	return strings.Join(append([]string{s.prefix}, keys...), "/")
}
