package state_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/state"
	"github.com/spf13/viper"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
// func Test(t *testing.T) { TestingT(t) }

type EngineSuite struct {
	ipvs        *ipvs.Ipvs
	service     *types.Service
	destination *types.Destination
	state       *state.State
	config      *config.BalancerConfig
}

var _ = Suite(&EngineSuite{})

func (s *EngineSuite) SetUpSuite(c *C) {
	logrus.SetOutput(ioutil.Discard)
	s.readConfig()
	c.Assert(s.config, Not(IsNil))

	s.service = &types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	s.destination = &types.Destination{
		Name:      "test",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *EngineSuite) SetUpTest(c *C) {
	eng, err := state.New(s.config)
	c.Assert(err, IsNil)

	s.state = eng

	go watchStateCh(eng)
}

func (s *EngineSuite) TearDownTest(c *C) {
	s.ipvs.Flush()
}

type MockSink struct {
	*bytes.Buffer
	cancel bool
}

func (m *MockSink) ID() string {
	return "Mock"
}

func (m *MockSink) Cancel() error {
	m.cancel = true
	return nil
}

func (m *MockSink) Close() error {
	return nil
}

func (s *EngineSuite) readConfig() {
	viper.SetConfigType("json")

	var sampleConfig = []byte(`
	{
		"provider":{
			"type": "none",
			"params": {
				"interface": "eth0",
				"vip-range": "192.168.0.0/28"
			}
		}
	}
	`)

	viper.ReadConfig(bytes.NewBuffer(sampleConfig))
	viper.Unmarshal(&s.config)
}

func makeLog(cmd *state.Command, c *C) *raft.Log {
	bytes, err := json.Marshal(cmd)
	c.Assert(err, IsNil)

	return &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  bytes,
	}
}

func watchStateCh(state *state.State) {
	for {
		errCh := <-state.ChangesCh()
		errCh <- nil
	}
}

func (s *EngineSuite) addService(c *C) {
	cmd := &state.Command{
		Op:      state.AddServiceOp,
		Service: s.service,
	}

	resp := s.state.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)
}

func (s *EngineSuite) delService(c *C) {
	cmd := &state.Command{
		Op:      state.DelServiceOp,
		Service: s.service,
	}

	resp := s.state.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)
}

func (s *EngineSuite) addDestination(c *C) {
	cmd := &state.Command{
		Op:          state.AddDestinationOp,
		Service:     s.service,
		Destination: s.destination,
	}

	resp := s.state.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)
}

func (s *EngineSuite) TestApplyAddService(c *C) {
	s.addService(c)

	c.Assert(s.state.Store.GetServices(), DeepEquals, []types.Service{*s.service})
}

func (s *EngineSuite) TestApplyDelService(c *C) {
	s.addService(c)
	s.delService(c)

	c.Assert(s.state.Store.GetServices(), DeepEquals, []types.Service{})
}

func (s *EngineSuite) TestApplyAddDestination(c *C) {
	s.addService(c)
	s.addDestination(c)

	dst, err := s.state.Store.GetDestination(s.destination.Name)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, s.destination)
}

func (s *EngineSuite) TestApplyDelDestination(c *C) {
	s.addService(c)
	s.addDestination(c)

	cmd := &state.Command{
		Op:          state.DelDestinationOp,
		Service:     s.service,
		Destination: s.destination,
	}

	resp := s.state.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)

	_, err := s.state.Store.GetDestination(s.destination.Name)
	c.Assert(err, Equals, types.ErrDestinationNotFound)
}

func (s *EngineSuite) TestSnapshotRestore(c *C) {
	s.addService(c)
	s.addDestination(c)

	snap, err := s.state.Snapshot()
	c.Assert(err, IsNil)
	defer snap.Release()

	buf := bytes.NewBuffer(nil)
	sink := &MockSink{buf, false}
	err = snap.Persist(sink)
	c.Assert(err, IsNil)

	eng, err := state.New(s.config)
	c.Assert(err, IsNil)
	go watchStateCh(eng)

	err = eng.Restore(sink)
	c.Assert(err, IsNil)

	// s.service.Destinations = []types.Destination{*s.destination}

	c.Assert(eng.Store.GetServices(), DeepEquals, []types.Service{*s.service})
}
