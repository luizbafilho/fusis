package api

import (
	"net/http/httptest"
	"os"
	"testing"

	"gopkg.in/check.v1"
)

type S struct {
	srv *httptest.Server
	bal *testBalancer
}

var _ = check.Suite(&S{})

func Test(t *testing.T) { check.TestingT(t) }

func (s *S) SetUpTest(c *check.C) {
	os.Unsetenv("FUSIS_ENV")
	s.bal = newTestBalancer()
	api := NewAPI(s.bal)
	s.srv = httptest.NewServer(api.router)
}

func (s *S) TearDownTest(c *check.C) {
	s.srv.Close()
}
