package engine_test

import (
	"bytes"
	"encoding/json"
	"testing"

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
}

var _ = Suite(&EngineSuite{})

func (s *EngineSuite) SetUpSuite(c *C) {
	s.readConfig()
}

func (s *EngineSuite) readConfig() {

	viper.SetConfigType("json") // or viper.SetConfigType("YAML")

	// any approach to require this configuration into your program.
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
		// fmt.Fatalf("err: %v", err)
	}

	return &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  bytes,
	}
}

func (s *EngineSuite) SetUpTest(c *C) {
	s.service = &ipvs.Service{
		Name:         "test",
		Host:         "10.0.1.1",
		Port:         80,
		Scheduler:    "lc",
		Protocol:     "tcp",
		Destinations: []ipvs.Destination{},
	}

	s.destination = &ipvs.Destination{
		Host:   "192.168.1.1",
		Port:   80,
		Mode:   "nat",
		Weight: 1,
	}
}

func (s *EngineSuite) TearDownTest(c *C) {
	s.ipvs.Flush()
}

func watchCommandCh(engine *engine.Engine) {
	<-engine.CommandCh
}

func (s *EngineSuite) TestApplyAddService(c *C) {
	cmd := &engine.Command{
		Op:      engine.AddServiceOp,
		Service: s.service,
	}

	eng, err := engine.New()
	c.Assert(err, IsNil)
	go watchCommandCh(eng)

	resp := eng.Apply(makeLog(cmd))
	if resp != nil {
		c.Fatalf("resp: %v", resp)
	}

	c.Assert(eng.State.GetServices(), DeepEquals, &[]ipvs.Service{*s.service})
	svcs, err := eng.Ipvs.GetServices()
	c.Assert(err, IsNil)

	c.Assert(svcs[0].Address.String(), DeepEquals, s.service.Host)
}
