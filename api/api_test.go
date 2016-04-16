package api_test

// import (
// 	"net/http"
// 	"testing"
// 	"time"
//
// 	"github.com/luizbafilho/fusis/api"
// 	. "github.com/luizbafilho/fusis/db"
// 	"github.com/luizbafilho/fusis/db/etcd"
//
// 	. "gopkg.in/check.v1"
// )
//
// // Hook up gocheck into the "go test" runner.
// func Test(t *testing.T) { TestingT(t) }
//
// type EtcdSuite struct {
// 	store       Store
// 	apiClient   api.Client
// 	service     Service
// 	destination Destination
// }
//
// var _ = Suite(&EtcdSuite{})
//
// func (s *EtcdSuite) SetUpSuite(c *C) {
// 	nodes := []string{"http://127.0.0.1:2379"}
// 	s.store = etcd.New(nodes, "fusis_test")
//
// 	s.apiClient = *api.NewClient("http://localhost:8000")
// }
//
// func (s *EtcdSuite) SetUpTest(c *C) {
// 	s.service = Service{
// 		Host:         "10.7.0.1",
// 		Port:         80,
// 		Scheduler:    "lc",
// 		Protocol:     "tcp",
// 		Destinations: []Destination{},
// 	}
//
// 	s.destination = Destination{
// 		Host:   "192.168.1.1",
// 		Port:   80,
// 		Mode:   "nat",
// 		Weight: 1,
// 	}
//
// 	flushStoreAndIpvs()
// }
//
// // func (s *EtcdSuite) TearDownTest(c *C) {
// // 	flushStoreAndIpvs()
// // }
//
// func (s *EtcdSuite) TestGestServices(c *C) {
// 	err := s.apiClient.UpsertService(s.service)
// 	c.Assert(err, IsNil)
// 	time.Sleep(time.Millisecond * 500)
//
// 	services, err := s.apiClient.GetServices()
// 	c.Assert(err, IsNil)
//
// 	c.Assert(*services, DeepEquals, []Service{s.service})
// }
//
// func (s *EtcdSuite) TestUpsertService(c *C) {
// 	err := s.apiClient.UpsertService(s.service)
// 	c.Assert(err, IsNil)
//
// 	svc, err := s.store.GetService(s.service.GetId())
// 	c.Assert(*svc, DeepEquals, s.service)
// }
//
// func (s *EtcdSuite) TestDeleteService(c *C) {
// 	err := s.store.UpsertService(s.service)
// 	c.Assert(err, IsNil)
//
// 	err = s.apiClient.DeleteService(s.service)
// 	c.Assert(err, IsNil)
//
// 	_, err = s.store.GetService(s.service.GetId())
// 	c.Assert(err, NotNil)
// }
//
// func (s *EtcdSuite) TestUpsertDestination(c *C) {
// 	err := s.store.UpsertService(s.service)
// 	c.Assert(err, IsNil)
//
// 	err = s.apiClient.UpsertDestination(s.service, s.destination)
// 	c.Assert(err, IsNil)
//
// 	destinations, err := s.store.GetDestinations(s.service)
// 	c.Assert(err, IsNil)
//
// 	c.Assert(*destinations, DeepEquals, []Destination{s.destination})
// }
//
// func (s *EtcdSuite) TestDeleteDestination(c *C) {
// 	err := s.store.UpsertService(s.service)
// 	c.Assert(err, IsNil)
//
// 	err = s.apiClient.UpsertDestination(s.service, s.destination)
// 	c.Assert(err, IsNil)
//
// 	err = s.apiClient.DeleteDestination(s.service, s.destination)
// 	c.Assert(err, IsNil)
//
// 	dsts, err := s.store.GetDestinations(s.service)
// 	c.Assert(*dsts, DeepEquals, []Destination{})
// }
//
// func flushStoreAndIpvs() {
// 	resp, err := http.Post("http://localhost:8000/flush", "", nil)
// 	if err != nil || resp.StatusCode != 200 {
// 		panic(err)
// 	}
// }
