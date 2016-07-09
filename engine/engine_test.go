package engine_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/spf13/viper"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type EngineSuite struct {
	ipvs        *ipvs.Ipvs
	service     *types.Service
	destination *types.Destination
	engine      *engine.Engine
	config      *config.BalancerConfig
}

var _ = Suite(&EngineSuite{})

func (s *EngineSuite) SetUpSuite(c *C) {
	logrus.SetOutput(ioutil.Discard)
	s.readConfig()
	c.Assert(s.config, Not(IsNil))

	s.service = &types.Service{
		Name:         "test",
		Host:         "10.0.1.1",
		Port:         80,
		Scheduler:    "lc",
		Protocol:     "tcp",
		Destinations: []types.Destination{},
	}

	s.destination = &types.Destination{
		Name:      "test",
		Host:      "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *EngineSuite) SetUpTest(c *C) {
	eng, err := engine.New(s.config)
	c.Assert(err, IsNil)

	s.engine = eng

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
				"vipRange": "192.168.0.0/28"
			}
		}
	}
	`)

	viper.ReadConfig(bytes.NewBuffer(sampleConfig))
	viper.Unmarshal(&s.config)
}

func makeLog(cmd *engine.Command, c *C) *raft.Log {
	bytes, err := json.Marshal(cmd)
	c.Assert(err, IsNil)

	return &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  bytes,
	}
}

func watchStateCh(engine *engine.Engine) {
	for {
		errCh := <-engine.StateCh
		errCh <- nil
	}
}

func (s *EngineSuite) addService(c *C) {
	cmd := &engine.Command{
		Op:      engine.AddServiceOp,
		Service: s.service,
	}

	resp := s.engine.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)
}

func (s *EngineSuite) delService(c *C) {
	cmd := &engine.Command{
		Op:      engine.DelServiceOp,
		Service: s.service,
	}

	resp := s.engine.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)
}

func (s *EngineSuite) addDestination(c *C) {
	cmd := &engine.Command{
		Op:          engine.AddDestinationOp,
		Service:     s.service,
		Destination: s.destination,
	}

	resp := s.engine.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)
}

func (s *EngineSuite) TestApplyAddService(c *C) {
	s.addService(c)

	c.Assert(s.engine.State.GetServices(), DeepEquals, []types.Service{*s.service})
}

func (s *EngineSuite) TestApplyDelService(c *C) {
	s.addService(c)
	s.delService(c)

	c.Assert(s.engine.State.GetServices(), DeepEquals, []types.Service{})
}

func (s *EngineSuite) TestApplyAddDestination(c *C) {
	s.addService(c)
	s.addDestination(c)

	dst, err := s.engine.State.GetDestination(s.destination.Name)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, s.destination)
}

func (s *EngineSuite) TestApplyDelDestination(c *C) {
	s.addService(c)
	s.addDestination(c)

	cmd := &engine.Command{
		Op:          engine.DelDestinationOp,
		Service:     s.service,
		Destination: s.destination,
	}

	resp := s.engine.Apply(makeLog(cmd, c))
	c.Assert(resp, IsNil)

	_, err := s.engine.State.GetDestination(s.destination.Name)
	c.Assert(err, Equals, types.ErrDestinationNotFound)
}

func (s *EngineSuite) TestSnapshotRestore(c *C) {
	s.addService(c)
	s.addDestination(c)

	snap, err := s.engine.Snapshot()
	c.Assert(err, IsNil)
	defer snap.Release()

	buf := bytes.NewBuffer(nil)
	sink := &MockSink{buf, false}
	err = snap.Persist(sink)
	c.Assert(err, IsNil)

	eng, err := engine.New(s.config)
	c.Assert(err, IsNil)
	go watchStateCh(eng)

	err = eng.Restore(sink)
	c.Assert(err, IsNil)

	s.service.Destinations = []types.Destination{*s.destination}

	c.Assert(eng.State.GetServices(), DeepEquals, []types.Service{*s.service})
}
