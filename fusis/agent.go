package fusis

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipvs"
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

func (a *Agent) Shutdown() {
	if err := a.serf.Leave(); err != nil {
		log.Fatalf("Graceful shutdown failed", err)
	}
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

func (a *Agent) Start() error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "agent"

	bindAddr, err := a.config.GetIpByInterface()
	if err != nil {
		panic(err)
	}

	conf.NodeName = a.config.Name

	conf.MemberlistConfig.BindAddr = bindAddr
	conf.EventCh = a.eventCh

	serf, err := serf.Create(conf)
	if err != nil {
		return err
	}

	a.serf = serf

	go a.handleEvents()
	return nil
}

func (a *Agent) handleEvents() {
	for {
		select {
		case e := <-a.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				memberEvent := e.(serf.MemberEvent)

				for _, m := range memberEvent.Members {
					if m.Name == a.serf.LocalMember().Name {
						a.broadcastToBalancers()
					}
				}
			default:
				log.Warnf("Fusis Agent: unhandled Serf Event: %#v", e)
			}
		}
	}
}

func (a *Agent) broadcastToBalancers() {
	host, err := a.config.GetIpByInterface()
	if err != nil {
		panic(err)
	}

	if a.config.Host != "" {
		host = a.config.Host
	}

	dst := ipvs.Destination{
		Name:      a.config.Name,
		Host:      host,
		Port:      a.config.Port,
		Weight:    1,
		Mode:      a.config.Mode,
		ServiceId: a.config.Service,
	}

	params := serf.QueryParam{
		FilterTags: map[string]string{"role": "balancer"},
	}

	payload, err := dst.ToJson()
	if err != nil {
		log.Errorf("Fusis Agent: Destination Marshaling failed: %v", err)
		return
	}

	log.Infof("Fusis Agent: broadcasting agent join to balancers. Host: %v", host)
	_, err = a.serf.Query("add-destination", payload, &params)
	if err != nil {
		log.Errorf("Fusis Agent: add-balancer event error: %v", err)
	}
}
