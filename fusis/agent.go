package fusis

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
)

type Agent struct {
	serf *serf.Serf
	// eventCh is used for Serf to deliver events on
	eventCh chan serf.Event
}

func NewAgent() (*Agent, error) {
	agent := &Agent{
		eventCh: make(chan serf.Event, 64),
	}

	return agent, nil
}

func (a *Agent) Join(existing []string, ignoreOld bool) (n int, err error) {
	log.Infof("Fusis Agent: joining: %v ignore: %v", existing, ignoreOld)
	n, err = a.serf.Join(existing, ignoreOld)
	if n > 0 {
		log.Infof("Fusis Agent: joined: %d nodes", n)
	}
	if err == nil {
		log.Warnf("Fusis Agent: error joining: %v", err)
	}
	return
}

func (a *Agent) Start(config Config) error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "agent"

	bindAddr, err := config.GetIpByInterface()
	if err != nil {
		panic(err)
	}

	conf.MemberlistConfig.BindAddr = bindAddr
	conf.EventCh = a.eventCh

	serf, err := serf.Create(conf)
	if err != nil {
		return err
	}

	a.serf = serf

	go a.eventLoop()
	return nil
}

func (a *Agent) eventLoop() {
	for {
		select {
		case e := <-a.eventCh:
			log.Infof("Fusis Agent: Received event: %s", e.String())
		}
	}
}
