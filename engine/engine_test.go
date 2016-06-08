package engine_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/spf13/viper"

	_ "github.com/luizbafilho/fusis/provider/none" // to intialize
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type EngineSuite struct {
	ipvs        *ipvs.Ipvs
	service     *ipvs.Service
	destination *ipvs.Destination
	engine      *engine.Engine
}

var _ = Suite(&EngineSuite{})

func (s *EngineSuite) SetUpSuite(c *C) {
	logrus.SetOutput(ioutil.Discard)
	s.readConfig()

	s.service = &ipvs.Service{
		Name:         "test",
		Host:         "10.0.1.1",
		Port:         80,
		Scheduler:    "lc",
		Protocol:     "tcp",
		Destinations: []ipvs.Destination{},
	}

	s.destination = &ipvs.Destination{
		Name:      "test",
		Host:      "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}

func (s *EngineSuite) SetUpTest(c *C) {
	eng, err := engine.New()
	c.Assert(err, IsNil)

	s.engine = eng

	go watchCommandCh(eng)
}

func (s *EngineSuite) TearDownTest(c *C) {
	s.ipvs.Flush()
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
	viper.Unmarshal(&config.Balancer)
}

func makeLog(cmd *engine.Command) *raft.Log {
	bytes, err := json.Marshal(cmd)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	return &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  bytes,
	}
}

func watchCommandCh(engine *engine.Engine) {
	for {
		<-engine.CommandCh
	}
}

func (s *EngineSuite) addService(c *C) {
	cmd := &engine.Command{
		Op:      engine.AddServiceOp,
		Service: s.service,
	}

	resp := s.engine.Apply(makeLog(cmd))
	if resp != nil {
		c.Fatalf("resp: %v", resp)
	}
}

func (s *EngineSuite) delService(c *C) {
	cmd := &engine.Command{
		Op:      engine.DelServiceOp,
		Service: s.service,
	}

	resp := s.engine.Apply(makeLog(cmd))
	if resp != nil {
		c.Fatalf("resp: %v", resp)
	}
}

func (s *EngineSuite) addDestination(c *C) {
	cmd := &engine.Command{
		Op:          engine.AddDestinationOp,
		Service:     s.service,
		Destination: s.destination,
	}

	resp := s.engine.Apply(makeLog(cmd))
	if resp != nil {
		c.Fatalf("resp: %v", resp)
	}
}

func (s *EngineSuite) TestApplyAddService(c *C) {
	s.addService(c)

	c.Assert(s.engine.State.GetServices(), DeepEquals, &[]ipvs.Service{*s.service})
	svcs, err := s.engine.Ipvs.GetServices()
	c.Assert(err, IsNil)

	c.Assert(len(svcs), Equals, 1)
	c.Assert(svcs[0].Address.String(), DeepEquals, s.service.Host)
}

func (s *EngineSuite) TestApplyDelService(c *C) {
	s.addService(c)
	s.delService(c)

	c.Assert(s.engine.State.GetServices(), DeepEquals, &[]ipvs.Service{})
	svcs, err := s.engine.Ipvs.GetServices()
	c.Assert(err, IsNil)

	c.Assert(len(svcs), Equals, 0)
}

func (s *EngineSuite) TestApplyAddDestination(c *C) {
	s.addService(c)
	s.addDestination(c)

	dst, err := s.engine.State.GetDestination(s.destination.Name)
	c.Assert(err, IsNil)

	c.Assert(dst, DeepEquals, s.destination)
	dests, err := s.engine.Ipvs.GetDestinations(s.service.ToIpvsService())
	c.Assert(err, IsNil)

	c.Assert(len(dests), Equals, 1)
	c.Assert(dests[0].Address.String(), DeepEquals, s.destination.Host)
}

func (s *EngineSuite) TestApplyDelDestination(c *C) {
	s.addService(c)
	s.addDestination(c)

	cmd := &engine.Command{
		Op:          engine.DelDestinationOp,
		Service:     s.service,
		Destination: s.destination,
	}

	resp := s.engine.Apply(makeLog(cmd))
	if resp != nil {
		c.Fatalf("resp: %v", resp)
	}

	dests, err := s.engine.Ipvs.GetDestinations(s.service.ToIpvsService())
	c.Assert(err, IsNil)

	c.Assert(len(dests), Equals, 0)
}
