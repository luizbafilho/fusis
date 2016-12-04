package store

import (
	"encoding/json"
	"net"
	"net/url"
	"strings"
	"sync"

	validator "gopkg.in/go-playground/validator.v9"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	kv "github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/types"
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
	GetServices() ([]types.Service, error)
	// GetService(serviceId string) (types.Service, error)
	AddService(svc *types.Service) error
	DeleteService(svc *types.Service) error
	SubscribeServices(ch chan []types.Service)
	WatchServices()

	GetDestinations() ([]types.Destination, error)
	AddDestination(svc *types.Service, dst *types.Destination) error
	DeleteDestination(svc *types.Service, dst *types.Destination) error
	SubscribeDestinations(ch chan []types.Destination)
	WatchDestinations()

	AddCheck(check types.CheckSpec) error
	DeleteCheck(check types.CheckSpec) error
	SubscribeChecks(ch chan []types.CheckSpec)
	WatchChecks()

	GetKV() kv.Store
}

var (
	ErrUnsupportedStore = errors.New("[store] Unsupported store")
)

type FusisStore struct {
	sync.Mutex
	kv kv.Store

	prefix string

	validate            *validator.Validate
	servicesChannels    []chan []types.Service
	destinationChannels []chan []types.Destination
	checksChannels      []chan []types.CheckSpec
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

	//Validating open connection
	_, err = net.Dial("tcp", u.Host)
	if err != nil {
		return nil, errors.Wrap(err, "Store connection failed. Make sure your store is up and running.")
	}

	kv, err := libkv.NewStore(
		kv.Backend(scheme),
		[]string{u.Host},
		nil,
	)
	if err != nil {
		kv.Close()
		return nil, errors.Wrap(err, "Cannot create store consul")
	}

	svcsChs := []chan []types.Service{}
	dstsChs := []chan []types.Destination{}
	checksChs := []chan []types.CheckSpec{}

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
	}

	go fusisStore.WatchServices()
	go fusisStore.WatchDestinations()
	go fusisStore.WatchChecks()

	return fusisStore, nil
}

func (s *FusisStore) GetKV() kv.Store {
	return s.kv
}

func (s *FusisStore) GetServices() ([]types.Service, error) {
	svcs := []types.Service{}
	entries, err := s.kv.List(s.key("services"))
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return svcs, nil
		}

		return nil, err
	}

	for _, pair := range entries {
		svc := types.Service{}
		if err := json.Unmarshal(pair.Value, &svc); err != nil {
			log.Error("[store] ", err)
		}

		svcs = append(svcs, svc)
	}

	return svcs, nil
}

func (s *FusisStore) GetDestinations() ([]types.Destination, error) {
	dsts := []types.Destination{}
	entries, err := s.kv.List("fusis/destinations")
	if err != nil {
		return nil, err
	}

	for _, pair := range entries {
		dst := types.Destination{}
		if err := json.Unmarshal(pair.Value, &dst); err != nil {
			log.Error("[store] ", err)
		}

		dsts = append(dsts, dst)
	}

	return dsts, nil
}

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
	if err := s.validateServiceNameUniqueness(svcKey); err != nil {
		return err
	}
	if err := s.validateServiceIpvsUniqueness(ipvsKey); err != nil {
		return err
	}

	// TODO: Make the persistence of service and ipvs id atomic.
	// Pesisting service
	value, err := json.Marshal(svc)
	if err != nil {
		return errors.Wrapf(err, "error marshaling service: %v", svc)
	}
	err = s.kv.Put(svcKey, value, nil)
	if err != nil {
		return errors.Wrapf(err, "error sending service to store: %v", svc)
	}

	// Persisting IPVS key. So it can be validated.
	err = s.kv.Put(ipvsKey, []byte("true"), nil)
	if err != nil {
		return errors.Wrapf(err, "error sending service ipvs id to store: %v", svc.IpvsId())
	}

	return nil
}

func (s *FusisStore) GetService(serviceId string) (*types.Service, error) {
	key := s.key("services", serviceId, "config")
	kvPair, err := s.kv.Get(key)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return nil, types.ErrServiceNotFound
		}
		return nil, errors.Wrap(err, "GetService from store failed")
	}

	svc := &types.Service{}
	if err := json.Unmarshal(kvPair.Value, svc); err != nil {
		return nil, errors.Wrap(err, "Unmarshal service from store failed")
	}

	return svc, nil
}

func (s *FusisStore) DeleteService(svc *types.Service) error {
	key := s.key("services", svc.GetId())
	err := s.kv.DeleteTree(key)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete service: %v", svc)
	}

	ipvsKey := s.key("ipvs-ids", "services", svc.IpvsId())
	err = s.kv.Delete(ipvsKey)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete service ipvs id: %s", ipvsKey)
	}

	// TODO: Delete all dependent destinations and checks

	return nil
}

func (s *FusisStore) SubscribeServices(updateCh chan []types.Service) {
	s.Lock()
	defer s.Unlock()
	s.servicesChannels = append(s.servicesChannels, updateCh)
}

func (s *FusisStore) WatchServices() {
	svcs := []types.Service{}

	stopCh := make(<-chan struct{})
	events, err := s.kv.WatchTree(s.key("services"), stopCh)
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

			s.Lock()
			for _, ch := range s.servicesChannels {
				log.Debug("[store] Sending message to service ch")
				ch <- svcs
			}
			s.Unlock()

			//Cleaning up services slice
			svcs = []types.Service{}
		}
	}
}

func (s *FusisStore) AddDestination(svc *types.Service, dst *types.Destination) error {
	dstKey := s.key("destinations", svc.GetId(), dst.GetId())
	ipvsKey := s.key("ipvs-ids", "destinations", dst.IpvsId())

	// Validating destination
	if err := s.validateDestination(dst); err != nil {
		return err
	}
	if err := s.validateDestinationNameUniqueness(dstKey); err != nil {
		return err
	}
	if err := s.validateDestinationIpvsUniqueness(ipvsKey); err != nil {
		return err
	}

	// Persisting destination
	value, err := json.Marshal(dst)
	if err != nil {
		return errors.Wrapf(err, "error marshaling destination: %v", dst)
	}
	err = s.kv.Put(dstKey, value, nil)
	if err != nil {
		return errors.Wrapf(err, "error sending destination to store: %v", dst)
	}
	log.Debugf("[store] Added destination: %s with key: %s", value, dstKey)

	// Persisting IPVS key. So it can be validated.
	err = s.kv.Put(ipvsKey, []byte("true"), nil)
	if err != nil {
		return errors.Wrapf(err, "error sending destination ipvs id to store: %v", dst.IpvsId())
	}

	return nil
}

func (s *FusisStore) DeleteDestination(svc *types.Service, dst *types.Destination) error {
	key := s.key("destinations", svc.GetId(), dst.GetId())

	err := s.kv.DeleteTree(key)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete destination: %v", dst)
	}
	log.Debugf("[store] Deleted destination: %s", key)

	ipvsKey := s.key("ipvs-ids", "destinations", dst.IpvsId())
	err = s.kv.Delete(ipvsKey)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete destination ipvs id: %s", ipvsKey)
	}

	return nil
}

func (s *FusisStore) SubscribeDestinations(updateCh chan []types.Destination) {
	s.Lock()
	defer s.Unlock()
	s.destinationChannels = append(s.destinationChannels, updateCh)
}

func (s *FusisStore) WatchDestinations() {
	dsts := []types.Destination{}

	stopCh := make(<-chan struct{})
	events, err := s.kv.WatchTree(s.key("destinations"), stopCh)
	if err != nil {
		errors.Wrap(err, "failed watching fusis/destinations")
	}

	for {
		select {
		case entries := <-events:
			log.Debug("[store] Destinations received")

			for _, pair := range entries {
				dst := types.Destination{}
				if err := json.Unmarshal(pair.Value, &dst); err != nil {
					errors.Wrap(err, "failed unmarshall of destinations")
				}
				log.Debugf("[store] Got destination: %#v", dst)

				dsts = append(dsts, dst)
			}

			s.Lock()
			for _, ch := range s.destinationChannels {
				ch <- dsts
			}
			s.Unlock()

			//Cleaning up destinations slice
			dsts = []types.Destination{}
		}
	}
}

func (s *FusisStore) SubscribeChecks(updateCh chan []types.CheckSpec) {
	s.Lock()
	defer s.Unlock()
	s.checksChannels = append(s.checksChannels, updateCh)
}

func (s *FusisStore) AddCheck(spec types.CheckSpec) error {
	key := s.key("checks", spec.ServiceID)

	value, err := json.Marshal(spec)
	if err != nil {
		return errors.Wrapf(err, "error marshaling CheckSpec: %#v", spec)
	}

	err = s.kv.Put(key, value, nil)
	if err != nil {
		return errors.Wrapf(err, "error sending CheckSpec to store: %#v", spec)
	}

	return nil
}

func (s *FusisStore) DeleteCheck(spec types.CheckSpec) error {
	key := s.key("checks", spec.ServiceID)

	err := s.kv.DeleteTree(key)
	if err != nil {
		return errors.Wrapf(err, "error trying to delete check: %#v", spec)
	}
	log.Debugf("[store] Deleted check: %s", key)

	return nil
}

func (s *FusisStore) WatchChecks() {
	specs := []types.CheckSpec{}

	stopCh := make(<-chan struct{})
	events, err := s.kv.WatchTree(s.key("checks"), stopCh)
	if err != nil {
		log.Error(errors.Wrap(err, "failed watching fusis/checks"))
	}

	for {
		select {
		case entries := <-events:
			log.Debug("[store] Checks received")

			for _, pair := range entries {
				spec := types.CheckSpec{}
				if err := json.Unmarshal(pair.Value, &spec); err != nil {
					log.Error(errors.Wrap(err, "failed unmarshall of checks"))
				}
				log.Debugf("[store] Got Check: %#v", spec)

				specs = append(specs, spec)
			}

			s.Lock()
			for _, ch := range s.checksChannels {
				ch <- specs
			}
			s.Unlock()

			//Cleaning up destinations slice
			specs = []types.CheckSpec{}
		}
	}
}

func (s *FusisStore) key(keys ...string) string {
	return strings.Join(append([]string{s.prefix}, keys...), "/")
}
