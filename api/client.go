package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/luizbafilho/fusis/types"
)

type Client struct {
	Addr       string
	HttpClient *http.Client
}

func NewClient(addr string) *Client {
	baseTimeout := 30 * time.Second
	fullTimeout := time.Minute
	return &Client{
		Addr: addr,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   baseTimeout,
					KeepAlive: baseTimeout,
				}).Dial,
				TLSHandshakeTimeout: baseTimeout,
				// Disabled http keep alive for more reliable dial timeouts.
				MaxIdleConnsPerHost: -1,
				DisableKeepAlives:   true,
			},
			Timeout: fullTimeout,
		},
	}
}

func (c *Client) GetServices() ([]types.Service, error) {
	svcs := []types.Service{}
	resp, err := c.HttpClient.Get(c.path("services"))
	if err != nil {
		return svcs, err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return svcs, err
	}

	err = decode(resp.Body, &svcs)
	if err != nil {
		return svcs, err
	}

	return svcs, nil
}

func (c *Client) GetService(id string) (*types.Service, error) {
	var svc *types.Service

	resp, err := c.HttpClient.Get(c.path("services", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return svc, err
	}

	err = decode(resp.Body, &svc)
	if err != nil {
		return svc, err
	}

	return svc, nil
}

func (c *Client) CreateService(svc types.Service) error {
	json, err := encode(svc)
	if err != nil {
		return err
	}
	resp, err := c.HttpClient.Post(c.path("services"), "application/json", json)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteService(id string) error {
	req, err := http.NewRequest("DELETE", c.path("services", id), nil)
	if err != nil {
		return err
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddDestination(dst types.Destination) (string, error) {
	json, err := encode(dst)
	if err != nil {
		return "", err
	}
	resp, err := c.HttpClient.Post(c.path("services", dst.ServiceId, "destinations"), "application/json", json)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var id string
	switch resp.StatusCode {
	case http.StatusNotFound:
		err = types.ErrServiceNotFound
	case http.StatusConflict:
		err = types.ErrDestinationConflict
	case http.StatusCreated:
		id = idFromLocation(resp)
	default:
		err = formatError(resp)
	}
	return id, err
}

func (c *Client) DeleteDestination(serviceId, destinationId string) error {
	req, err := http.NewRequest("DELETE", c.path("services", serviceId, "destinations", destinationId), nil)
	if err != nil {
		return err
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotFound:
		err = types.ErrDestinationNotFound
	case http.StatusNoContent:
	default:
		err = formatError(resp)
	}
	return err
}

func encode(obj interface{}) (io.Reader, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func decode(body io.Reader, obj interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return fmt.Errorf("unable to unmarshal body %q: %s", string(data), err)
	}
	return nil
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	var errMsg struct {
		Error interface{}
	}

	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, &errMsg)
		if err != nil {
			return err
		}

		switch r.StatusCode {
		case http.StatusNotFound:
			err = types.ErrNotFound(errMsg.Error.(string))
		case http.StatusUnprocessableEntity:
			err = types.ErrValidation{Errors: errMsg.Error.(map[string]string)}
		default:
			err = errors.New(errMsg.Error.(string))
		}

		return err

	}

	return fmt.Errorf("Request failed. Empty response from the server. Status Code: %v. ", r.StatusCode)
}

func formatError(resp *http.Response) error {
	body, _ := ioutil.ReadAll(resp.Body)
	return fmt.Errorf("Request failed. Status Code: %v. Body: %q", resp.StatusCode, string(body))
}

func (c Client) path(paths ...string) string {
	return strings.Join(append([]string{strings.TrimRight(c.Addr, "/")}, paths...), "/")
}

func idFromLocation(resp *http.Response) string {
	parts := strings.Split(resp.Header.Get("Location"), "/")
	return parts[len(parts)-1]
}

// func (r *ErrorResponse) Error() string {
// 	var msg interface{}
// 	msg = r.Errs
// 	if r.Err != "" {
// 		msg = r.Err
// 	}
//
// 	return fmt.Sprintf("%v %v: %d %v", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, msg)
// }
