package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luizbafilho/fusis/ipvs"
	"gopkg.in/check.v1"
)

type S struct{}

var _ = check.Suite(&S{})

func Test(t *testing.T) { check.TestingT(t) }

func (s *S) TestNewClient(c *check.C) {
	cli := NewClient("myaddr")
	c.Assert(cli, check.NotNil)
	c.Assert(cli.Addr, check.Equals, "myaddr")
	c.Assert(cli.HttpClient, check.NotNil)
}

func (s *S) TestClientGetServices(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Write([]byte(`[{"id": "id1", "name": "name1"}, {"id": "id2", "name": "name2"}]`))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, []*ipvs.Service{
		{Id: "id1", Name: "name1"},
		{Id: "id2", Name: "name2"},
	})
	c.Assert(req.Method, check.Equals, "GET")
	c.Assert(req.URL.Path, check.Equals, "/services")
}

func (s *S) TestClientGetServicesEmpty(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, []*ipvs.Service{})
}

func (s *S) TestClientGetServicesError(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("some error"))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"some error\"")
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetServicesUnparseable(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid"))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.ErrorMatches, "unable to unmarshal body \"invalid\": invalid character 'i' looking for beginning of value")
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetService(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Write([]byte(`{"id": "id1", "name": "name1"}`))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetService("id1")
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, &ipvs.Service{
		Id: "id1", Name: "name1",
	})
	c.Assert(req.Method, check.Equals, "GET")
	c.Assert(req.URL.Path, check.Equals, "/services/id1")
}

func (s *S) TestClientGetServiceNotFound(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetService("id1")
	c.Assert(err, check.Equals, ErrNoSuchService)
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetServiceError(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("some error"))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetService("id1")
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"some error\"")
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetServiceUnparseable(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid"))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	result, err := cli.GetService("id1")
	c.Assert(err, check.ErrorMatches, "unable to unmarshal body \"invalid\": invalid character 'i' looking for beginning of value")
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientCreateService(c *check.C) {
	var (
		req  *http.Request
		body []byte
		err  error
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		body, err = ioutil.ReadAll(r.Body)
		c.Assert(err, check.IsNil)
		w.Header().Set("Location", "/services/mysvc")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	id, err := cli.CreateService(ipvs.Service{Name: "name1"})
	c.Assert(err, check.IsNil)
	c.Assert(id, check.Equals, "mysvc")
	c.Assert(req.Method, check.Equals, "POST")
	c.Assert(req.URL.Path, check.Equals, "/services")
	c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/json")
	var result ipvs.Service
	err = json.Unmarshal(body, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, ipvs.Service{Name: "name1"})
}

func (s *S) TestClientCreateServiceInvalidStatus(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	id, err := cli.CreateService(ipvs.Service{Name: "name1"})
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 200. Body: \"\"")
	c.Assert(id, check.Equals, "")
}

func (s *S) TestClientDeleteService(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	err := cli.DeleteService("id1")
	c.Assert(err, check.IsNil)
	c.Assert(req.Method, check.Equals, "DELETE")
	c.Assert(req.URL.Path, check.Equals, "/services/id1")
}

func (s *S) TestClientDeleteServiceInvalidStatus(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	err := cli.DeleteService("id1")
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"\"")
}

func (s *S) TestClientAddDestination(c *check.C) {
	var (
		req  *http.Request
		body []byte
		err  error
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		body, err = ioutil.ReadAll(r.Body)
		c.Assert(err, check.IsNil)
		w.Header().Set("Location", "/services/svid1/destinations/mydst")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	id, err := cli.AddDestination(ipvs.Destination{ServiceId: "svid1"})
	c.Assert(err, check.IsNil)
	c.Assert(id, check.Equals, "mydst")
	c.Assert(req.Method, check.Equals, "POST")
	c.Assert(req.URL.Path, check.Equals, "/services/svid1/destinations")
	c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/json")
	var result ipvs.Destination
	err = json.Unmarshal(body, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, ipvs.Destination{ServiceId: "svid1"})
}

func (s *S) TestClientAddDestinationInvalidStatus(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	id, err := cli.AddDestination(ipvs.Destination{ServiceId: "svid1"})
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 200. Body: \"\"")
	c.Assert(id, check.Equals, "")
}

func (s *S) TestClientDeleteDestination(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	err := cli.DeleteDestination("svid1", "dstid1")
	c.Assert(err, check.IsNil)
	c.Assert(req.Method, check.Equals, "DELETE")
	c.Assert(req.URL.Path, check.Equals, "/services/svid1/destinations/dstid1")
}

func (s *S) TestClientDeleteDestinationInvalidStatus(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	cli := NewClient(srv.URL)
	err := cli.DeleteDestination("svid1", "dstid1")
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"\"")
}
