package fusis

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/engine"

	"github.com/hashicorp/serf/serf"
)

type Balancer struct {
	serf    *serf.Serf
	eventCh chan serf.Event
}

func NewBalancer() (*Balancer, error) {
	balancer := &Balancer{
		eventCh: make(chan serf.Event, 64),
	}

	return balancer, nil
}

func (b *Balancer) Start(config Config) error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "balancer"

	bindAddr, err := config.GetIpByInterface()
	if err != nil {
		panic(err)
	}

	conf.MemberlistConfig.BindAddr = bindAddr
	conf.EventCh = b.eventCh

	serf, err := serf.Create(conf)
	if err != nil {
		return err
	}

	b.serf = serf

	go b.handleEvents()
	return nil
}

func (b *Balancer) handleEvents() {
	for {
		select {
		case e := <-b.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				memberEvent := e.(serf.MemberEvent)
				fmt.Println("=====>> ", memberEvent.Members)
			case serf.EventQuery:
				query := e.(*serf.Query)
				b.handleQuery(query)
			default:
				log.Warnf("Fusis Balancer: unhandled Serf Event: %#v", e)
			}
		}
	}
}

func (b *Balancer) handleQuery(query *serf.Query) {
	name := query.Name
	payload := query.Payload

	var dst engine.Destination
	err := json.Unmarshal(payload, &dst)
	if err != nil {
		log.Errorf("Fusis Balancer: Unable to Unmarshal: %v", payload)
	}

	fmt.Println("=====> Recebendo Dst name:", name)
	fmt.Println("=====> Recebendo Dst: payload", dst)
}
