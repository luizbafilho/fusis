package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/luizbafilho/fusis/ipvs"
	"gopkg.in/check.v1"
)

type testBalancer struct {
	services []ipvs.Service
	ids      map[string]int
	dests    map[string][]int
}

func newTestBalancer() *testBalancer {
	return &testBalancer{
		ids:   make(map[string]int),
		dests: make(map[string][]int),
	}
}

func (b *testBalancer) GetServices() []ipvs.Service {
	return b.services
}

func (b *testBalancer) AddService(srv *ipvs.Service) error {
	_, exists := b.ids[srv.Id]
	if exists {
		return errors.New("conflict?")
	}
	b.services = append(b.services, *srv)
	b.ids[srv.Id] = len(b.services) - 1
	return nil
}

func (b *testBalancer) GetService(id string) (*ipvs.Service, error) {
	idx, exists := b.ids[id]
	if !exists {
		return nil, ipvs.ErrNotFound
	}
	return &b.services[idx], nil
}

func (b *testBalancer) DeleteService(id string) error {
	idx, exists := b.ids[id]
	if !exists {
		return ipvs.ErrNotFound
	}
	delete(b.ids, id)
	b.services = append(b.services[:idx], b.services[idx+1:]...)
	return nil
}

func (b *testBalancer) AddDestination(srv *ipvs.Service, dest *ipvs.Destination) error {
	idx, exists := b.ids[srv.Id]
	if !exists {
		return ipvs.ErrNotFound
	}
	_, exists = b.dests[dest.Id]
	if exists {
		return errors.New("conflict?")
	}
	srv = &b.services[idx]
	srv.Destinations = append(srv.Destinations, *dest)
	b.dests[dest.Id] = []int{idx, len(srv.Destinations) - 1}
	return nil
}

func (b *testBalancer) GetDestination(id string) (*ipvs.Destination, error) {
	indexes, exists := b.dests[id]
	if !exists {
		return nil, ipvs.ErrNotFound
	}
	return &b.services[indexes[0]].Destinations[indexes[1]], nil
}

func (b *testBalancer) DeleteDestination(dest *ipvs.Destination) error {
	indexes, exists := b.dests[dest.Id]
	if !exists {
		return ipvs.ErrNotFound
	}
	srv := &b.services[indexes[0]]
	srv.Destinations = append(srv.Destinations[:indexes[1]], srv.Destinations[indexes[1]:]...)
	return nil
}

func (s *S) TestNewAPI(c *check.C) {
	b := testBalancer{}
	api := NewAPI(&b)
	c.Assert(api.balancer, check.Equals, &b)
	c.Assert(api.env, check.Equals, "development")
}

func (s *S) TestServiceList(c *check.C) {
	err := s.bal.AddService(&ipvs.Service{Id: "myservice"})
	c.Assert(err, check.IsNil)
	resp, err := http.Get(s.srv.URL + "/services")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result []ipvs.Service
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, []ipvs.Service{{Id: "myservice"}})
}

func (s *S) TestServiceListEmpty(c *check.C) {
	resp, err := http.Get(s.srv.URL + "/services")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNoContent)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	c.Assert(data, check.DeepEquals, []byte{})
}

func (s *S) TestServiceGet(c *check.C) {
	err := s.bal.AddService(&ipvs.Service{Id: "myservice"})
	c.Assert(err, check.IsNil)
	resp, err := http.Get(s.srv.URL + "/services/myservice")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result ipvs.Service
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, ipvs.Service{Id: "myservice"})
}

func (s *S) TestServiceGetNotFound(c *check.C) {
	err := s.bal.AddService(&ipvs.Service{Id: "myservice"})
	c.Assert(err, check.IsNil)
	resp, err := http.Get(s.srv.URL + "/services/myservice2")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNotFound)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result map[string]string
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, map[string]string{"error": "Service not found"})
}

func (s *S) TestServiceCreate(c *check.C) {
	body := strings.NewReader(`{"id": "mysrv", "name": "ahoy", "port": 1040, "protocol": "tcp", "scheduler": "rr"}`)
	resp, err := http.Post(s.srv.URL+"/services", "application/json", body)
	c.Assert(err, check.IsNil)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result ipvs.Service
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, ipvs.Service{
		Id:           "mysrv",
		Name:         "ahoy",
		Port:         1040,
		Protocol:     "tcp",
		Scheduler:    "rr",
		Destinations: []ipvs.Destination{},
	})
	c.Assert(resp.StatusCode, check.Equals, http.StatusCreated)
	c.Assert(resp.Header.Get("Location"), check.Matches, `/services/mysrv`)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
}

func (s *S) TestServiceCreateValidationError(c *check.C) {
	body := strings.NewReader(`{"id": "mysrv"}`)
	resp, err := http.Post(s.srv.URL+"/services", "application/json", body)
	c.Assert(err, check.IsNil)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result map[string]map[string]string
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, map[string]map[string]string{
		"errors": {
			"Name":      "non zero value required",
			"Port":      "non zero value required",
			"Protocol":  "non zero value required",
			"Scheduler": "non zero value required",
		},
	})
	c.Assert(resp.StatusCode, check.Equals, http.StatusBadRequest)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
}

func (s *S) TestServiceDelete(c *check.C) {
	err := s.bal.AddService(&ipvs.Service{Id: "myservice"})
	c.Assert(err, check.IsNil)
	req, err := http.NewRequest("DELETE", s.srv.URL+"/services/myservice", nil)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNoContent)
	_, err = s.bal.GetService("myservice")
	c.Assert(err, check.Equals, ipvs.ErrNotFound)
}

func (s *S) TestServiceDeleteNotFound(c *check.C) {
	err := s.bal.AddService(&ipvs.Service{Id: "myservice"})
	c.Assert(err, check.IsNil)
	req, err := http.NewRequest("DELETE", s.srv.URL+"/services/myservice2", nil)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNotFound)
}
