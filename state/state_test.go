package state_test

import (
	"testing"

	"github.com/luizbafilho/fusis/state"
	"github.com/stretchr/testify/suite"
)

type StateTestSuite struct {
	suite.Suite
	state state.State
}

func (suite *StateTestSuite) SetupTest() {
	// config :=
	// suite.state =
}

func (suite *StateTestSuite) TestState() {
	// assert.Equal(suite.T(), 5, suite.VariableThatShouldStartAtFive)
}

func TestStateTestSuite(t *testing.T) {
	suite.Run(t, new(StateTestSuite))
}

// Hook up gocheck into the "go test" runner.
// func Test(t *testing.T) { TestingT(t) }

// type EngineSuite struct {
// 	ipvs        *ipvs.Ipvs
// 	service     *types.Service
// 	destination *types.Destination
// 	state       *state.State
// 	config      *config.BalancerConfig
// }
//
// var _ = Suite(&EngineSuite{})
//
// func (s *EngineSuite) SetUpSuite(c *C) {
// 	logrus.SetOutput(ioutil.Discard)
// 	s.readConfig()
// 	c.Assert(s.config, Not(IsNil))
//
// 	s.service = &types.Service{
// 		Name:      "test",
// 		Address:   "10.0.1.1",
// 		Port:      80,
// 		Scheduler: "lc",
// 		Protocol:  "tcp",
// 	}
//
// 	s.destination = &types.Destination{
// 		Name:      "test",
// 		Address:   "192.168.1.1",
// 		Port:      80,
// 		Mode:      "nat",
// 		Weight:    1,
// 		ServiceId: "test",
// 	}
// }
//
// func (s *EngineSuite) SetUpTest(c *C) {
// 	eng, err := state.New(s.config)
// 	c.Assert(err, IsNil)
//
// 	s.state = eng
//
// 	go watchStateCh(eng)
// }
//
// func (s *EngineSuite) TearDownTest(c *C) {
// 	s.ipvs.Flush()
// }
//
// type MockSink struct {
// 	*bytes.Buffer
// 	cancel bool
// }
//
// func (m *MockSink) ID() string {
// 	return "Mock"
// }
//
// func (m *MockSink) Cancel() error {
// 	m.cancel = true
// 	return nil
// }
//
// func (m *MockSink) Close() error {
// 	return nil
// }
//
// func (s *EngineSuite) readConfig() {
// 	viper.SetConfigType("json")
//
// 	var sampleConfig = []byte(`
// 	{
// 		"provider":{
// 			"type": "none",
// 			"params": {
// 				"interface": "eth0",
// 				"vip-range": "192.168.0.0/28"
// 			}
// 		}
// 	}
// 	`)
//
// 	viper.ReadConfig(bytes.NewBuffer(sampleConfig))
// 	viper.Unmarshal(&s.config)
// }
//
// func watchStateCh(state *state.State) {
// 	for {
// 		errCh := <-state.ChangesCh()
// 		errCh <- nil
// 	}
// }
//
// func (s *StateSuite) TestGetService(c *C) {
// 	s.state.AddService(s.service)
// 	s.state.AddDestination(s.destination)
//
// 	svcs := s.state.GetServices()
// 	// s.service.Destinations = []types.Destination{*s.destination}
// 	c.Assert(svcs[0], DeepEquals, *s.service)
//
// 	svc, err := s.state.GetService(s.service.Name)
// 	c.Assert(err, IsNil)
// 	c.Assert(svc, DeepEquals, s.service)
//
// 	_, err = s.state.GetService("unknown")
// 	c.Assert(err, Equals, types.ErrServiceNotFound)
// }
//
// func (s *StateSuite) TestAddService(c *C) {
// 	s.state.AddService(s.service)
//
// 	service, err := s.state.GetService(s.service.Name)
// 	c.Assert(err, IsNil)
// 	c.Assert(service, DeepEquals, s.service)
// }
//
// func (s *StateSuite) TestDelService(c *C) {
// 	s.state.AddService(s.service)
// 	s.state.DeleteService(s.service)
//
// 	services := s.state.GetServices()
// 	c.Assert(len(services), Equals, 0)
//
// 	_, err := s.state.GetService(s.service.Name)
// 	c.Assert(err, Equals, types.ErrServiceNotFound)
// }
//
// func (s *StateSuite) TestAddDestination(c *C) {
// 	s.state.AddService(s.service)
// 	s.state.AddDestination(s.destination)
//
// 	dst, err := s.state.GetDestination(s.destination.Name)
// 	c.Assert(err, IsNil)
// 	c.Assert(dst, DeepEquals, s.destination)
// }
//
// func (s *StateSuite) TestDelDestination(c *C) {
// 	s.state.AddService(s.service)
// 	s.state.AddDestination(s.destination)
// 	s.state.DeleteDestination(s.destination)
//
// 	_, err := s.state.GetDestination(s.destination.Name)
// 	c.Assert(err, DeepEquals, types.ErrDestinationNotFound)
// }
