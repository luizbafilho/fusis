package etcd

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
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

func (s Etcd) GetServices() (*[]store.Service, error) {
	key := fmt.Sprintf("/fusis/services")

	r, err := s.client.Get(context.Background(), key, &client.GetOptions{Recursive: true, Sort: true})

	var services []store.Service

	for _, n := range r.Node.Nodes {
		conf := n.Nodes[0]
		var s store.Service

		err := getValueFromJson(conf.Value, &s)
		if err != nil {
			return nil, err
		}

		services = append(services, s)
	}

	return &services, err
}

//TODO: Avoid insertion of destination as values on the service
func (s Etcd) UpsertService(svc store.Service) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s/conf", svc.Host, svc.Port, svc.Protocol)

	value, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	_, err = s.client.Set(context.Background(), key, string(value), nil)

	return err
}

func (s Etcd) DeleteService(svc store.Service) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s", svc.Host, svc.Port, svc.Protocol)

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Services does not exists.")
	}

	_, err := s.client.Delete(context.Background(), key, &client.DeleteOptions{Recursive: true})

	return err
}

func (s Etcd) UpsertDestination(svc store.Service, dst store.Destination) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s/servers/%s-%v", svc.Host, svc.Port, svc.Protocol, dst.Host, dst.Port)

	value, err := json.Marshal(dst)
	if err != nil {
		return err
	}

	_, err = s.client.Set(context.Background(), key, string(value), nil)

	return err
}

func (s Etcd) DeleteDestination(svc store.Service, dst store.Destination) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s/servers/%s-%v", svc.Host, svc.Port, svc.Protocol, dst.Host, dst.Port)

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Destination does not exists.")
	}

	_, err := s.client.Delete(context.Background(), key, &client.DeleteOptions{})

	return err
}

func (s Etcd) Subscribe(changes chan interface{}) error {
	w := s.client.Watcher("/fusis/services", &client.WatcherOptions{AfterIndex: 0, Recursive: true})

	for {
		response, err := w.Next(context.Background())
		if err != nil {
			return err
		}
		event, _ := processChange(response)

		changes <- event
	}
}

type MatcherFn func(*client.Response) (interface{}, error)

// Dispatches etcd key changes changes to the etcd to the matching functions
func processChange(response *client.Response) (interface{}, error) {
	matchers := []MatcherFn{
		processServiceChange,
		processDestinationChange,
	}

	for _, matcher := range matchers {
		v, err := matcher(response)
		if v != nil || err != nil {
			return v, err
		}
	}
	return nil, nil
}

func processServiceChange(r *client.Response) (interface{}, error) {
	out := regexp.MustCompile(`/fusis/services/(.*)-(\d+)-(tcp|udp)(/conf)?$`).FindStringSubmatch(r.Node.Key)

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

func processDestinationChange(r *client.Response) (interface{}, error) {
	out := regexp.MustCompile(`/fusis/services/(.*)-(\d+)-(tcp|udp)/servers/(.*)-(\d+)`).FindStringSubmatch(r.Node.Key)

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
