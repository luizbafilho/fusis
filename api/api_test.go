package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/luizbafilho/fusis/api/types"
	"gopkg.in/check.v1"
)

type testBalancer struct {
	services []types.Service
	ids      map[string]int
	dests    map[string][]int
}

func newTestBalancer() *testBalancer {
	return &testBalancer{
		ids:   make(map[string]int),
		dests: make(map[string][]int),
	}
}

func (b *testBalancer) GetServices() []types.Service {
	return b.services
}

func (b *testBalancer) AddService(srv *types.Service) error {
	_, exists := b.ids[srv.Name]
	if exists {
		return errors.New("conflict?")
	}
	b.services = append(b.services, *srv)
	b.ids[srv.Name] = len(b.services) - 1
	return nil
}

func (b *testBalancer) GetService(id string) (*types.Service, error) {
	idx, exists := b.ids[id]
	if !exists {
		return nil, types.ErrServiceNotFound
	}
	return &b.services[idx], nil
}

func (b *testBalancer) DeleteService(id string) error {
	idx, exists := b.ids[id]
	if !exists {
		return types.ErrServiceNotFound
	}
	delete(b.ids, id)
	b.services = append(b.services[:idx], b.services[idx+1:]...)
	return nil
}

func (b *testBalancer) AddDestination(srv *types.Service, dest *types.Destination) error {
	idx, exists := b.ids[srv.Name]
	if !exists {
		return types.ErrServiceNotFound
	}
	_, exists = b.dests[dest.Name]
	if exists {
		return errors.New("conflict?")
	}
	srv = &b.services[idx]
	srv.Destinations = append(srv.Destinations, *dest)
	b.dests[dest.Name] = []int{idx, len(srv.Destinations) - 1}
	return nil
}

func (b *testBalancer) GetDestination(id string) (*types.Destination, error) {
	indexes, exists := b.dests[id]
	if !exists {
		return nil, types.ErrDestinationNotFound
	}
	return &b.services[indexes[0]].Destinations[indexes[1]], nil
}

func (b *testBalancer) DeleteDestination(dest *types.Destination) error {
	indexes, exists := b.dests[dest.Name]
	if !exists {
		return types.ErrDestinationNotFound
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
	err := s.bal.AddService(&types.Service{Name: "myservice"})
	c.Assert(err, check.IsNil)
	resp, err := http.Get(s.srv.URL + "/services")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result []types.Service
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, []types.Service{{Name: "myservice"}})
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
	err := s.bal.AddService(&types.Service{Name: "myservice"})
	c.Assert(err, check.IsNil)
	resp, err := http.Get(s.srv.URL + "/services/myservice")
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result types.Service
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, types.Service{Name: "myservice"})
}

func (s *S) TestServiceGetNotFound(c *check.C) {
	err := s.bal.AddService(&types.Service{Name: "myservice"})
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
	c.Assert(result, check.DeepEquals, map[string]string{"error": "service not found"})
}

func (s *S) TestServiceCreate(c *check.C) {
	body := strings.NewReader(`{"name": "ahoy", "port": 1040, "protocol": "tcp", "scheduler": "rr"}`)
	resp, err := http.Post(s.srv.URL+"/services", "application/json", body)
	c.Assert(err, check.IsNil)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result types.Service
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, types.Service{
		Name:         "ahoy",
		Port:         1040,
		Protocol:     "tcp",
		Scheduler:    "rr",
		Destinations: []types.Destination{},
	})
	c.Assert(resp.StatusCode, check.Equals, http.StatusCreated)
	c.Assert(resp.Header.Get("Location"), check.Matches, `/services/ahoy`)
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
	err := s.bal.AddService(&types.Service{Name: "myservice"})
	c.Assert(err, check.IsNil)
	req, err := http.NewRequest("DELETE", s.srv.URL+"/services/myservice", nil)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNoContent)
	_, err = s.bal.GetService("myservice")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
}

func (s *S) TestServiceDeleteNotFound(c *check.C) {
	err := s.bal.AddService(&types.Service{Name: "myservice"})
	c.Assert(err, check.IsNil)
	req, err := http.NewRequest("DELETE", s.srv.URL+"/services/myservice2", nil)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNotFound)
}

func (s *S) TestDestinationCreate(c *check.C) {
	err := s.bal.AddService(&types.Service{Name: "myservice"})
	c.Assert(err, check.IsNil)
	body := strings.NewReader(`{"name": "myname", "host": "myhost", "port": 1234}`)
	req, err := http.NewRequest("POST", s.srv.URL+"/services/myservice/destinations", body)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result types.Destination
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, types.Destination{
		Name:      "myname",
		Host:      "myhost",
		Port:      1234,
		Weight:    1,
		Mode:      "route",
		ServiceId: "myservice",
	})
	c.Assert(resp.StatusCode, check.Equals, http.StatusCreated)
	c.Assert(resp.Header.Get("Location"), check.Matches, `/services/myservice/destinations/myname`)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
}

func (s *S) TestDestinationCreateValidationError(c *check.C) {
	err := s.bal.AddService(&types.Service{Name: "myservice"})
	c.Assert(err, check.IsNil)
	body := strings.NewReader(`{"id": "mydest"}`)
	req, err := http.NewRequest("POST", s.srv.URL+"/services/myservice/destinations", body)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result map[string]map[string]string
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, map[string]map[string]string{
		"errors": {
			"Name": "non zero value required",
			"Port": "non zero value required",
			"Host": "non zero value required",
		},
	})
	c.Assert(resp.StatusCode, check.Equals, http.StatusBadRequest)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
}

func (s *S) TestDestinationCreateServiceNotFound(c *check.C) {
	body := strings.NewReader(`{"id": "mydest"}`)
	req, err := http.NewRequest("POST", s.srv.URL+"/services/myservice/destinations", body)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, check.IsNil)
	var result map[string]string
	err = json.Unmarshal(data, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, map[string]string{
		"error": "service not found",
	})
	c.Assert(resp.StatusCode, check.Equals, http.StatusNotFound)
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, "application/json; charset=utf-8")
}

func (s *S) TestDestinationDelete(c *check.C) {
	srv := &types.Service{Name: "myservice"}
	err := s.bal.AddService(srv)
	c.Assert(err, check.IsNil)
	dst := &types.Destination{
		Name:      "mydest",
		ServiceId: "myservice",
	}
	err = s.bal.AddDestination(srv, dst)
	c.Assert(err, check.IsNil)
	req, err := http.NewRequest("DELETE", s.srv.URL+"/services/myservice/destinations/mydest", nil)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNoContent)
}

func (s *S) TestDestinationDeleteNotFound(c *check.C) {
	srv := &types.Service{Name: "myservice"}
	err := s.bal.AddService(srv)
	c.Assert(err, check.IsNil)
	req, err := http.NewRequest("DELETE", s.srv.URL+"/services/myservice/destinations/mydest", nil)
	c.Assert(err, check.IsNil)
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode, check.Equals, http.StatusNotFound)
}
