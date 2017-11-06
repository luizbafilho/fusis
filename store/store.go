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
	GetService(name string) (*types.Service, error)
	AddService(svc *types.Service) error
	DeleteService(svc *types.Service) error

	GetDestinations(svc *types.Service) ([]types.Destination, error)
	AddDestination(svc *types.Service, dst *types.Destination) error
	DeleteDestination(svc *types.Service, dst *types.Destination) error

	AddCheck(check types.CheckSpec) error
	GetCheck(serviceID string) (*types.CheckSpec, error)
	DeleteCheck(check types.CheckSpec) error

	AddWatcher(ch chan state.State)
	Watch()

	Close() error
}

type FusisStore struct {
	sync.Mutex
	kv *clientv3.Client

	prefix string

	validate      *validator.Validate
	watchChannels []chan state.State
}

func New(config *config.BalancerConfig) (Store, error) {
	kv, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(config.EtcdEndpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "[store] connection to etcd failed")
	}

	watchChs := []chan state.State{}

	validate := validator.New()
	// Registering custom validations
	validate.RegisterValidation("protocols", validateValues(types.Protocols))
	validate.RegisterValidation("schedulers", validateValues(types.Schedulers))

	fusisStore := &FusisStore{
		kv:            kv,
		prefix:        config.StorePrefix,
		validate:      validate,
		watchChannels: watchChs,
	}

	return fusisStore, nil
}

func (s *FusisStore) Close() error {
	return s.kv.Close()
}

func (s *FusisStore) AddWatcher(ch chan state.State) {
	s.Lock()
	s.watchChannels = append(s.watchChannels, ch)
	s.Unlock()
}

func (s *FusisStore) Watch() {
	ticker := time.Tick(1 * time.Second)
	notify := false

	svcCh := s.kv.Watch(context.TODO(), s.key("services"), clientv3.WithPrefix())
	dstCh := s.kv.Watch(context.TODO(), s.key("destinations"), clientv3.WithPrefix())
	for {
		select {
		case <-svcCh:
			notify = true
		case <-dstCh:
			notify = true
		case <-ticker:
			if !notify {
				continue
			}

			notify = false
			fstate, err := s.GetState()
			if err != nil {
				logrus.Error(errors.Wrap(err, "[store] couldn't get state"))
			}

			for _, ch := range s.watchChannels {
				ch <- fstate
			}
		}
	}
}

func (s *FusisStore) GetState() (state.State, error) {
	fstate, _ := state.New()

	opts := append([]clientv3.OpOption{}, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	svcGet := clientv3.OpGet(s.key("services"), opts...)
	dstGet := clientv3.OpGet(s.key("destinations"), opts...)
	checkGet := clientv3.OpGet(s.key("checks"), opts...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := s.kv.Txn(ctx).Then(svcGet, dstGet, checkGet).Commit()
	cancel()
	if err != nil {
		return nil, errors.Wrap(err, "[store] Get data from etcd failed")
	}

	for _, r := range resp.Responses {
		for _, pair := range r.GetResponseRange().Kvs {
			key := string(pair.Key)
			switch {
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
			case strings.Contains(key, s.key("checks")):
				spec := types.CheckSpec{}
				if err := json.Unmarshal(pair.Value, &spec); err != nil {
					return nil, errors.Wrapf(err, "[store] Unmarshal %s failed", pair.Value)
				}
				fstate.AddCheck(spec)
			}
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

func (s *FusisStore) GetService(name string) (*types.Service, error) {
	resp, err := s.kv.Get(context.TODO(), s.key("services", name), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "[store] Get services failed")
	}

	if len(resp.Kvs) == 0 {
		return nil, types.ErrServiceNotFound
	}

	kv := resp.Kvs[0]
	svc := &types.Service{}
	if err := json.Unmarshal(kv.Value, &svc); err != nil {
		return nil, errors.Wrapf(err, "[store] unmarshal %s failed", kv.Value)
	}

	return svc, nil
}

func (s *FusisStore) GetDestinations(svc *types.Service) ([]types.Destination, error) {
	dsts := []types.Destination{}
	resp, err := s.kv.Get(context.TODO(), s.key("destinations", svc.GetId()), clientv3.WithPrefix())
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
	svcKey := s.key("services", svc.GetId())
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
		return types.ErrValidation{Type: "service", Errors: map[string]string{"ipvs": "service must be unique"}}
	}

	return nil
}

func (s *FusisStore) DeleteService(svc *types.Service) error {
	// Deleting service
	svcKey := s.key("services", svc.GetId())
	ipvsSvcKey := s.key("ipvs-ids", "services", svc.IpvsId())
	svcDel := clientv3.OpDelete(svcKey, clientv3.WithPrefix())
	ipvsSvcDel := clientv3.OpDelete(ipvsSvcKey, clientv3.WithPrefix())

	svcCmp := clientv3.Compare(clientv3.Version(svcKey), ">", 0)
	ipvsSvcCmp := clientv3.Compare(clientv3.Version(ipvsSvcKey), ">", 0)

	// Deleting destinations
	dstKey := s.key("destinations", svc.GetId())
	ipvsDstKey := s.key("ipvs-ids", "destinations", svc.GetId())
	dstDel := clientv3.OpDelete(dstKey, clientv3.WithPrefix())
	ipvsDstDel := clientv3.OpDelete(ipvsDstKey, clientv3.WithPrefix())

	resp, err := s.kv.Txn(context.TODO()).
		If(svcCmp, ipvsSvcCmp).
		Then(svcDel, ipvsSvcDel, dstDel, ipvsDstDel).
		Commit()
	if err != nil {
		return errors.Wrapf(err, "[store] Error deleting service from Etcd: %v", svc)
	}

	if resp.Succeeded == false {
		return types.ErrServiceNotFound
	}

	return nil
}

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
		return types.ErrValidation{Type: "destination", Errors: map[string]string{"ipvs": "destination must be unique"}}
	}

	return nil
}

func (s *FusisStore) DeleteDestination(svc *types.Service, dst *types.Destination) error {
	dstKey := s.key("destinations", svc.GetId(), dst.GetId())
	ipvsKey := s.key("ipvs-ids", "destinations", svc.GetId(), dst.IpvsId())

	dstDel := clientv3.OpDelete(dstKey, clientv3.WithPrefix())
	ipvsDel := clientv3.OpDelete(ipvsKey, clientv3.WithPrefix())

	dstCmp := clientv3.Compare(clientv3.Version(dstKey), ">", 0)
	ipvsCmp := clientv3.Compare(clientv3.Version(ipvsKey), ">", 0)

	resp, err := s.kv.Txn(context.TODO()).
		If(dstCmp, ipvsCmp).
		Then(dstDel, ipvsDel).
		Commit()
	if err != nil {
		return errors.Wrapf(err, "[store] Error deleting destination from Etcd: %v", svc)
	}

	// resp.Succeeded means the compare clause was true
	if resp.Succeeded == false {
		return types.ErrDestinationNotFound
	}

	return nil
}

func (s *FusisStore) AddCheck(spec types.CheckSpec) error {
	checkKey := s.key("checks", spec.ServiceID)

	// Persisting destination
	value, err := json.Marshal(spec)
	if err != nil {
		return errors.Wrapf(err, "[store] error marshaling spec: %#v", spec)
	}

	_, err = s.kv.Put(context.TODO(), checkKey, string(value))
	if err != nil {
		return errors.Wrapf(err, "[store] Error sending check to Etcd: %#v", spec)
	}

	return nil
}

func (s *FusisStore) GetCheck(serviceID string) (*types.CheckSpec, error) {
	checkKey := s.key("checks", serviceID)

	resp, err := s.kv.Get(context.TODO(), checkKey, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "[store] Error getting check from Etcd: %#v", checkKey)
	}

	if len(resp.Kvs) == 0 {
		return nil, types.ErrCheckNotFound
	}

	kv := resp.Kvs[0]
	check := &types.CheckSpec{}
	if err := json.Unmarshal(kv.Value, &check); err != nil {
		return nil, errors.Wrapf(err, "[store] unmarshal %s failed", kv.Value)
	}

	return check, nil
}

func (s *FusisStore) DeleteCheck(spec types.CheckSpec) error {
	checkKey := s.key("checks", spec.ServiceID)

	_, err := s.kv.Delete(context.TODO(), checkKey)
	if err != nil {
		return errors.Wrapf(err, "[store] Error deleting check from Etcd: %#v", spec)
	}

	return nil
}

func (s *FusisStore) key(keys ...string) string {
	lastIndex := len(keys) - 1
	keys[lastIndex] = keys[lastIndex] + "/"
	return strings.Join(append([]string{s.prefix}, keys...), "/")
}
