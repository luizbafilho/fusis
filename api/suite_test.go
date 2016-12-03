package api_test

// import (
// 	"os"
// 	"testing"
//
// 	"github.com/luizbafilho/fusis/api"
// 	apiTesting "github.com/luizbafilho/fusis/api/testing"
// 	"gopkg.in/check.v1"
// )
//
// type S struct {
// 	srv *apiTesting.FakeFusisServer
// 	bal api.Balancer
// }
//
// var _ = check.Suite(&S{})
//
// func Test(t *testing.T) { check.TestingT(t) }
//
// func (s *S) SetUpTest(c *check.C) {
// 	os.Unsetenv("FUSIS_ENV")
// 	s.srv = apiTesting.NewFakeFusisServer()
// 	s.bal = s.srv.Balancer
// }
//
// func (s *S) TearDownTest(c *check.C) {
// 	s.srv.Close()
// }
