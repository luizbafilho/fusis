package cluster

import (
	"fmt"

	"github.com/hashicorp/memberlist"
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
	fmt.Printf("[INFO] agent: joining: %v ignore: %v\n", existing, ignoreOld)
	n, err = a.serf.Join(existing, ignoreOld)
	if n > 0 {
		fmt.Printf("[INFO] fusis agent: joined: %d nodes\n", n)
	}
	if err != nil {
		fmt.Printf("[WARN] fusis agent: error joining: %v\n", err)
	}
	return
}

func (a *Agent) Start(bindAddr string) error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "agent"

	conf.MemberlistConfig = memberlist.DefaultLANConfig()
	conf.MemberlistConfig.BindAddr = bindAddr
	conf.MemberlistConfig.BindPort = 7946
	conf.RejoinAfterLeave = true
	conf.EventCh = a.eventCh

	fmt.Println("bind adrres ===>>", conf.MemberlistConfig.BindAddr)

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
			fmt.Printf("[INFO] fusis agent: Received event: %s", e.String())
		}
	}
}
