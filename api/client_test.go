package api_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/types"
	"gopkg.in/check.v1"
)

func (s *S) TestNewClient(c *check.C) {
	cli := api.NewClient("myaddr")
	c.Assert(cli, check.NotNil)
	c.Assert(cli.Addr, check.Equals, "myaddr")
	c.Assert(cli.HttpClient, check.NotNil)
}

func (s *S) TestClientGetServices(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Write([]byte(`[{"name": "name1"}, {"name": "name2"}]`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, []*types.Service{
		{Name: "name1"},
		{Name: "name2"},
	})
	c.Assert(req.Method, check.Equals, "GET")
	c.Assert(req.URL.Path, check.Equals, "/services")
}

func (s *S) TestClientGetServicesEmpty(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, []*types.Service{})
}

func (s *S) TestClientGetServicesError(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("some error"))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"some error\"")
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetServicesUnparseable(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid"))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetServices()
	c.Assert(err, check.ErrorMatches, "unable to unmarshal body \"invalid\": invalid character 'i' looking for beginning of value")
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetService(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Write([]byte(`{"name": "name1"}`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetService("name1")
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, &types.Service{
		Name: "name1",
	})
	c.Assert(req.Method, check.Equals, "GET")
	c.Assert(req.URL.Path, check.Equals, "/services/name1")
}

func (s *S) TestClientGetServiceNotFound(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetService("id1")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetServiceError(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("some error"))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetService("id1")
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"some error\"")
	c.Assert(result, check.IsNil)
}

func (s *S) TestClientGetServiceUnparseable(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid"))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
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
	cli := api.NewClient(srv.URL)
	id, err := cli.CreateService(types.Service{Name: "name1"})
	c.Assert(err, check.IsNil)
	c.Assert(id, check.Equals, "mysvc")
	c.Assert(req.Method, check.Equals, "POST")
	c.Assert(req.URL.Path, check.Equals, "/services")
	c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/json")
	var result types.Service
	err = json.Unmarshal(body, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, types.Service{Name: "name1"})
}

func (s *S) TestClientCreateServiceConflict(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	id, err := cli.CreateService(types.Service{Name: "name1"})
	c.Assert(err, check.Equals, types.ErrServiceAlreadyExists)
	c.Assert(id, check.Equals, "")
}

func (s *S) TestClientCreateServiceInvalidStatus(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	id, err := cli.CreateService(types.Service{Name: "name1"})
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 200. Body: \"\"")
	c.Assert(id, check.Equals, "")
}

func (s *S) TestClientDeleteService(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
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
	cli := api.NewClient(srv.URL)
	err := cli.DeleteService("id1")
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"\"")
}

func (s *S) TestClientDeleteServiceNotFound(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.DeleteService("id1")
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
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
	cli := api.NewClient(srv.URL)
	id, err := cli.AddDestination(types.Destination{ServiceId: "svid1"})
	c.Assert(err, check.IsNil)
	c.Assert(id, check.Equals, "mydst")
	c.Assert(req.Method, check.Equals, "POST")
	c.Assert(req.URL.Path, check.Equals, "/services/svid1/destinations")
	c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/json")
	var result types.Destination
	err = json.Unmarshal(body, &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, types.Destination{ServiceId: "svid1"})
}

func (s *S) TestClientAddDestinationInvalidStatus(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	id, err := cli.AddDestination(types.Destination{ServiceId: "svid1"})
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 200. Body: \"\"")
	c.Assert(id, check.Equals, "")
}

func (s *S) TestClientAddDestinationNotFound(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	id, err := cli.AddDestination(types.Destination{ServiceId: "svid1"})
	c.Assert(err, check.Equals, types.ErrServiceNotFound)
	c.Assert(id, check.Equals, "")
}

func (s *S) TestClientAddDestinationConflict(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	id, err := cli.AddDestination(types.Destination{ServiceId: "svid1"})
	c.Assert(err, check.Equals, types.ErrDestinationAlreadyExists)
	c.Assert(id, check.Equals, "")
}

func (s *S) TestClientDeleteDestination(c *check.C) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
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
	cli := api.NewClient(srv.URL)
	err := cli.DeleteDestination("svid1", "dstid1")
	c.Assert(err, check.ErrorMatches, "Request failed. Status Code: 500. Body: \"\"")
}

func (s *S) TestClientDeleteDestinationNotFound(c *check.C) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.DeleteDestination("svid1", "dstid1")
	c.Assert(err, check.Equals, types.ErrDestinationNotFound)
}
