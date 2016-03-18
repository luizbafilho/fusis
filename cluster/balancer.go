package cluster

import (
	"fmt"

	"github.com/hashicorp/memberlist"
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

func (b *Balancer) Start(bindAddr string) error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "balancer"

	conf.MemberlistConfig = memberlist.DefaultLANConfig()
	conf.MemberlistConfig.BindAddr = bindAddr
	conf.MemberlistConfig.BindPort = 7946
	conf.RejoinAfterLeave = true
	conf.EventCh = b.eventCh

	fmt.Println("bind adrres ===>>", conf.MemberlistConfig.BindAddr)

	serf, err := serf.Create(conf)

	if err != nil {
		return err
	}

	b.serf = serf

	go b.eventLoop()
	return nil
}

func (b *Balancer) eventLoop() {
	for {
		select {
		case e := <-b.eventCh:
			fmt.Printf("[INFO] fusis balancer: Received event: %s", e.String())
		}
	}
}
