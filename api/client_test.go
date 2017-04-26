package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	cli := api.NewClient("myaddr")
	assert.NotNil(t, cli)
	assert.Equal(t, cli.Addr, "myaddr")
	assert.NotNil(t, cli.HttpClient)
}

func TestClientGetServices(t *testing.T) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Write([]byte(`[{"name": "name1"}, {"name": "name2"}]`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetServices()
	assert.Nil(t, err)
	assert.Equal(t, []types.Service{
		{Name: "name1"},
		{Name: "name2"},
	}, result)

	assert.Equal(t, req.Method, "GET")
	assert.Equal(t, req.URL.Path, "/services")
}

func TestClientGetServicesEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetServices()
	assert.Nil(t, err)
	assert.Equal(t, []types.Service{}, result)
}

func TestClientGetServicesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "any error"}`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	_, err := cli.GetServices()
	assert.Equal(t, err.Error(), "any error")
}

func TestClientGetService(t *testing.T) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.Write([]byte(`{"name": "name1"}`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	result, err := cli.GetService("name1")
	assert.Nil(t, err)
	assert.Equal(t, &types.Service{Name: "name1"}, result)

	assert.Equal(t, req.Method, "GET")
	assert.Equal(t, req.URL.Path, "/services/name1")
}

//
func TestClientGetServiceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "service not found"}`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	_, err := cli.GetService("id1")
	assert.Equal(t, err.Error(), "service not found")
}

func TestClientCreateService(t *testing.T) {
	var (
		req  *http.Request
		body []byte
		err  error
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		body, err = ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		w.Header().Set("Location", "/services/name1")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	cli := api.NewClient(srv.URL)
	err = cli.CreateService(types.Service{Name: "name1"})

	assert.Nil(t, err)
	assert.Equal(t, req.Method, "POST")
	assert.Equal(t, req.URL.Path, "/services")
	assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
}

func TestClientCreateService_Validation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.CreateService(types.Service{Name: "name1"})
	assert.NotNil(t, err)
}

func TestClientDeleteService(t *testing.T) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.DeleteService("id1")
	assert.Nil(t, err)
	assert.Equal(t, req.Method, "DELETE")
	assert.Equal(t, req.URL.Path, "/services/id1")
}

func TestClientDeleteService_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "service not found"}`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.DeleteService("id1")
	assert.Equal(t, err.Error(), "service not found")
}

func TestClientAddDestination(t *testing.T) {
	var (
		req  *http.Request
		body []byte
		err  error
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		body, err = ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		w.Header().Set("Location", "/services/svid1/destinations/mydst")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err = cli.AddDestination(types.Destination{ServiceId: "svid1"})
	assert.Nil(t, err)
	assert.Equal(t, req.Method, "POST")
	assert.Equal(t, req.URL.Path, "/services/svid1/destinations")
	assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
}

func TestClientAddDestination_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "service not found"}`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.AddDestination(types.Destination{ServiceId: "svid1"})
	assert.Equal(t, err.Error(), "service not found")
}

func TestClientDeleteDestination(t *testing.T) {
	var req *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.DeleteDestination("svid1", "dstid1")
	assert.Nil(t, err)
	assert.Equal(t, req.Method, "DELETE")
	assert.Equal(t, req.URL.Path, "/services/svid1/destinations/dstid1")
}

func TestClientDeleteDestinationNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "destination not found"}`))
	}))
	defer srv.Close()
	cli := api.NewClient(srv.URL)
	err := cli.DeleteDestination("svid1", "dstid1")
	assert.Equal(t, err.Error(), "destination not found")
}
