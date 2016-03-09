package etcd

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
	"github.com/luizbafilho/fusis/store"
)

type Etcd struct {
	client  client.KeysAPI
	etcdKey string
}

func New(addrs []string, etcdKey string) store.Store {
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
		client:  client.NewKeysAPI(c),
		etcdKey: etcdKey,
	}

	return store
}

func (s Etcd) GetService(serviceId string) (*store.Service, error) {
	key := s.path("services", serviceId, "conf")

	r, err := s.client.Get(context.Background(), key, nil)
	if err != nil {
		return nil, err
	}

	var service store.Service
	err = getValueFromJson(r.Node.Value, &service)
	if err != nil {
		return nil, err
	}

	destinations, err := s.GetDestinations(service)
	if err != nil {
		return nil, err
	}

	service.Destinations = *destinations

	return &service, nil
}

func (s Etcd) GetServices() (*[]store.Service, error) {
	key := s.path("services")

	services := []store.Service{}

	r, err := s.client.Get(context.Background(), key, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		if err.(client.Error).Code == client.ErrorCodeKeyNotFound {
			return &services, nil
		}

		return nil, err
	}

	for _, n := range r.Node.Nodes {
		conf := n.Nodes[0]
		svc := store.Service{}

		err := getValueFromJson(conf.Value, &svc)
		if err != nil {
			return nil, err
		}

		destinations, err := s.GetDestinations(svc)
		if err != nil {
			return nil, err
		}

		svc.Destinations = *destinations

		services = append(services, svc)
	}

	return &services, err
}

//TODO: Avoid insertion of destination as values on the service
func (s Etcd) UpsertService(svc store.Service) error {
	key := s.path("services", svc.GetId(), "conf")

	value, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	_, err = s.client.Set(context.Background(), key, string(value), nil)

	return err
}

func (s Etcd) DeleteService(svc store.Service) error {
	key := s.path("services", svc.GetId())

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Services does not exists.")
	}

	_, err := s.client.Delete(context.Background(), key, &client.DeleteOptions{Recursive: true})

	return err
}

func (s Etcd) GetDestinations(svc store.Service) (*[]store.Destination, error) {
	key := s.path("services", svc.GetId(), "destinations")

	destinations := []store.Destination{}

	r, err := s.client.Get(context.Background(), key, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		if err.(client.Error).Code == client.ErrorCodeKeyNotFound {
			return &destinations, nil
		}

		return nil, err
	}

	for _, node := range r.Node.Nodes {
		var d store.Destination

		err := getValueFromJson(node.Value, &d)
		if err != nil {
			return nil, err
		}

		destinations = append(destinations, d)
	}

	return &destinations, err
}

func (s Etcd) UpsertDestination(svc store.Service, dst store.Destination) error {
	key := s.path("services", svc.GetId(), "conf")

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Services does not exists.")
	}

	key = s.path("services", svc.GetId(), "destinations", dst.GetId())

	value, err := json.Marshal(dst)
	if err != nil {
		return err
	}

	_, err = s.client.Set(context.Background(), key, string(value), nil)

	return err
}

func (s Etcd) DeleteDestination(svc store.Service, dst store.Destination) error {
	key := s.path("services", svc.GetId(), "destinations", dst.GetId())

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Destination does not exists.")
	}

	_, err := s.client.Delete(context.Background(), key, &client.DeleteOptions{})

	return err
}

func (s Etcd) Flush() error {
	exists, _ := s.keyExists(s.etcdKey)
	if !exists {
		return fmt.Errorf("Key: %v does not exists.", s.etcdKey)
	}

	_, err := s.client.Delete(context.Background(), s.etcdKey, &client.DeleteOptions{Recursive: true})

	return err
}

func (s Etcd) Subscribe(changes chan interface{}) error {
	key := s.path("services")
	w := s.client.Watcher(key, &client.WatcherOptions{AfterIndex: 0, Recursive: true})

	for {
		response, err := w.Next(context.Background())
		if err != nil {
			return err
		}
		event, _ := s.processChange(response)

		changes <- event
	}
}

type MatcherFn func(*client.Response) (interface{}, error)

// Dispatches etcd key changes changes to the etcd to the matching functions
func (s Etcd) processChange(response *client.Response) (interface{}, error) {
	matchers := []MatcherFn{
		s.processServiceChange,
		s.processDestinationChange,
	}

	for _, matcher := range matchers {
		v, err := matcher(response)
		if v != nil || err != nil {
			return v, err
		}
	}
	return nil, nil
}

func (s Etcd) processServiceChange(r *client.Response) (interface{}, error) {
	regex := fmt.Sprintf("/%s/services/(.*)-(\\d+)-(tcp|udp)(/conf)?$", s.etcdKey)
	out := regexp.MustCompile(regex).FindStringSubmatch(r.Node.Key)

	if len(out) != 5 {
		return nil, nil
	}

	var serviceRequest store.Service

	switch r.Action {
	case store.CreateEvent, store.SetEvent:
		getValueFromJson(r.Node.Value, &serviceRequest)

		return store.ServiceEvent{
			Action:  store.SetEvent,
			Service: serviceRequest,
		}, nil

	case store.DeleteEvent:
		getServiceFromRegexMatch(&serviceRequest, out)

		return store.ServiceEvent{
			Action:  store.DeleteEvent,
			Service: serviceRequest,
		}, nil
	}

	return nil, fmt.Errorf("unsupported action on the rate: %s", r.Action)
}

func (s Etcd) processDestinationChange(r *client.Response) (interface{}, error) {
	regex := fmt.Sprintf("/%s/services/(.*)-(\\d+)-(tcp|udp)/destinations/(.*)-(\\d+)", s.etcdKey)

	out := regexp.MustCompile(regex).FindStringSubmatch(r.Node.Key)
	if len(out) != 6 {
		return nil, nil
	}

	var dstRequest store.Destination
	var svcRequest store.Service

	getServiceFromRegexMatch(&svcRequest, out)

	switch r.Action {
	case store.CreateEvent, store.SetEvent:
		getValueFromJson(r.Node.Value, &dstRequest)

		return store.DestinationEvent{
			Action:      store.SetEvent,
			Service:     svcRequest,
			Destination: dstRequest,
		}, nil

	case store.DeleteEvent:
		getDestinationFromRegexMatch(&dstRequest, out)

		return store.DestinationEvent{
			Action:      store.DeleteEvent,
			Service:     svcRequest,
			Destination: dstRequest,
		}, nil
	}

	return nil, fmt.Errorf("unsupported action on the rate: %s", r.Action)
}

func getValueFromJson(value string, v interface{}) error {
	return json.Unmarshal([]byte(value), v)
}

func getServiceFromRegexMatch(service *store.Service, out []string) {
	service.Host = out[1]
	service.Protocol = out[3]

	u, _ := strconv.ParseUint(out[2], 10, 64)
	service.Port = uint16(u)
}

func getDestinationFromRegexMatch(dst *store.Destination, out []string) {
	dst.Host = out[4]

	u, _ := strconv.ParseUint(out[5], 10, 64)
	dst.Port = uint16(u)
}

func (s Etcd) keyExists(key string) (bool, error) {
	_, err := s.client.Get(context.Background(), key, nil)
	if err != nil {
		if keyNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// keyNotFound checks on the error returned by the KeysAPI
// to verify if the key exists in the store or not
func keyNotFound(err error) bool {
	if err != nil {
		if etcdError, ok := err.(client.Error); ok {
			if etcdError.Code == client.ErrorCodeKeyNotFound ||
				etcdError.Code == client.ErrorCodeNotFile ||
				etcdError.Code == client.ErrorCodeNotDir {
				return true
			}
		}
	}
	return false
}

func (s Etcd) path(keys ...string) string {
	return strings.Join(append([]string{s.etcdKey}, keys...), "/")
}
