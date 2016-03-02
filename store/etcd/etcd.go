package etcd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
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

func (s Etcd) AddService(svc store.ServiceRequest) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s/conf", svc.Host, svc.Port, svc.Protocol)

	exists, _ := s.keyExists(key)
	if exists {
		return fmt.Errorf("Services already exists")
	}

	value, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	_, err = s.client.Set(context.Background(), key, string(value), nil)

	return err
}

func (s Etcd) UpdateService(svc store.ServiceRequest) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s/conf", svc.Host, svc.Port, svc.Protocol)

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Services does not exists.")
	}

	value, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	_, err = s.client.Update(context.Background(), key, string(value))

	return err
}

func (s Etcd) DeleteService(svc store.ServiceRequest) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s", svc.Host, svc.Port, svc.Protocol)

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Services does not exists.")
	}

	_, err := s.client.Delete(context.Background(), key, &client.DeleteOptions{Recursive: true})

	return err
}

func (s Etcd) AddDestination(svc store.ServiceRequest, dst store.DestinationRequest) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s/servers/%s-%v", svc.Host, svc.Port, svc.Protocol, dst.Host, dst.Port)

	exists, _ := s.keyExists(key)
	if exists {
		return fmt.Errorf("Destination already exists")
	}

	value, err := json.Marshal(dst)
	if err != nil {
		return err
	}

	_, err = s.client.Set(context.Background(), key, string(value), nil)

	return err
}

func (s Etcd) UpdateDestination(svc store.ServiceRequest, dst store.DestinationRequest) error {
	key := fmt.Sprintf("/fusis/services/%s-%v-%s/servers/%s-%v", svc.Host, svc.Port, svc.Protocol, dst.Host, dst.Port)

	exists, _ := s.keyExists(key)
	if !exists {
		return fmt.Errorf("Destination does not exists.")
	}

	value, err := json.Marshal(dst)
	if err != nil {
		return err
	}

	_, err = s.client.Update(context.Background(), key, string(value))

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

	var serviceRequest store.ServiceRequest

	switch r.Action {
	case store.CreateEvent, store.SetEvent:
		getValueFromJson(r.Node.Value, &serviceRequest)

		return store.ServiceEvent{
			Action:  store.CreateEvent,
			Service: serviceRequest,
		}, nil

	case store.UpdateEvent:
		getValueFromJson(r.Node.Value, &serviceRequest)

		return store.ServiceEvent{
			Action:  store.UpdateEvent,
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

	var dstRequest store.DestinationRequest
	var svcRequest store.ServiceRequest

	getServiceFromRegexMatch(&svcRequest, out)

	switch r.Action {
	case store.CreateEvent, store.SetEvent:
		getValueFromJson(r.Node.Value, &dstRequest)

		return store.DestinationEvent{
			Action:      store.CreateEvent,
			Service:     svcRequest,
			Destination: dstRequest,
		}, nil

	case store.UpdateEvent:
		getValueFromJson(r.Node.Value, &dstRequest)

		return store.DestinationEvent{
			Action:      store.UpdateEvent,
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

func getServiceFromRegexMatch(service *store.ServiceRequest, out []string) {
	service.Host = net.ParseIP(out[1])

	u, _ := strconv.ParseUint(out[2], 10, 64)
	service.Port = uint16(u)

	service.Protocol.UnmarshalJSON([]byte(out[3]))
}

func getDestinationFromRegexMatch(dst *store.DestinationRequest, out []string) {
	dst.Host = net.ParseIP(out[4])

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
