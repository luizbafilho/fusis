package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/luizbafilho/fusis/ipvs"
)

type Client struct {
	Addr string
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr}
}

func (c *Client) GetServices() ([]*Service, error) {
	resp, err := http.Get(c.path("services"))
	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}

	services := []*Service{}
	err = decode(resp.Body, &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (c *Client) GetService(id string) (*Service, error) {
	resp, err := http.Get(c.path("services", id))
	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}

	var svc Service
	err = decode(resp.Body, &svc)
	if err != nil {
		return nil, err
	}

	return &svc, nil
}

func (c *Client) CreateService(svc Service) error {
	json, err := encode(svc)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.path("services"), "application/json", json)

	if err != nil || resp.StatusCode != 200 {
		return formatError(resp)
	}

	return nil
}

func (c *Client) DeleteService(id string) error {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", c.path("services", id), nil)
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != 200 {
		return formatError(resp)
	}

	return nil
}

func (c *Client) AddDestination(dst Destination) error {
	json, err := encode(dst)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.path("services", dst.ServiceId, "destinations"), "application/json", json)
	if err != nil || resp.StatusCode != 200 {
		return formatError(resp)
	}
	return nil
}

func (c *Client) DeleteDestination(dst Destination) error {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", c.path("services", dst.ServiceId, "destinations", dst.GetId()), nil)
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != 200 {
		return formatError(resp)
	}

	return nil
}

func encode(obj interface{}) (io.Reader, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func decode(body io.Reader, obj interface{}) error {
	decoder := json.NewDecoder(body)
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

func formatError(resp *http.Response) error {
	var body string
	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		body = string(b)
	}
	return fmt.Errorf("Request failed. Status Code: %v. Body: %v", resp.StatusCode, body)
}

func (c Client) path(paths ...string) string {
	return strings.Join(append([]string{c.Addr}, paths...), "/")
}
