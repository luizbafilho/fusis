package fusis

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
)

type Agent struct {
	serf *serf.Serf
	// eventCh is used for Serf to deliver events on
	eventCh chan serf.Event
	config  *config.AgentConfig
}

func NewAgent(config *config.AgentConfig) (*Agent, error) {
	log.Infof("Fusis Agent: Config ==> %+v", config)
	agent := &Agent{
		eventCh: make(chan serf.Event, 64),
		config:  config,
	}

	return agent, nil
}

func (a *Agent) Start() error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "agent"
	conf.Tags["info"] = a.getInfo()

	bindAddr, err := a.config.GetIpByInterface()
	if err != nil {
		log.Fatal(err)
	}

	conf.NodeName = a.config.Name

	conf.MemberlistConfig.BindAddr = bindAddr
	conf.EventCh = a.eventCh

	serf, err := serf.Create(conf)
	if err != nil {
		return err
	}

	a.serf = serf

	return nil
}

func (a *Agent) getInfo() string {
	host, err := a.config.GetIpByInterface()
	if err != nil {
		log.Fatal("Unable to get agent host address", err)
	}

	if a.config.Host != "" {
		host = a.config.Host
	}

	dst := types.Destination{
		Name:      a.config.Name,
		Host:      host,
		Port:      a.config.Port,
		Weight:    1,
		Mode:      a.config.Mode,
		ServiceId: a.config.Service,
	}

	payload, err := json.Marshal(dst)
	if err != nil {
		log.Fatal("Unable to marshal agent info", err)
	}

	return string(payload)
}

func (a *Agent) Join(existing []string, ignoreOld bool) (n int, err error) {
	log.Infof("Fusis Agent: joining: %v ignore: %v", existing, ignoreOld)
	n, err = a.serf.Join(existing, ignoreOld)
	if n > 0 {
		log.Infof("Fusis Agent: joined: %d nodes", n)
	}
	if err != nil {
		log.Warnf("Fusis Agent: error joining: %v", err)
	}
	return
}

func (a *Agent) Shutdown() {
	if err := a.serf.Leave(); err != nil {
		log.Fatalf("Graceful shutdown failed: %s", err)
	}
}
