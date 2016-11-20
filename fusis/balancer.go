package fusis

import (
	"fmt"
	"io"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/bgp"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/ipam"
	"github.com/luizbafilho/fusis/iptables"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/metrics"
	fusis_net "github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/store"
	"github.com/luizbafilho/fusis/vip"

	"github.com/hashicorp/logutils"
)

// Balancer represents the Load Balancer
type Balancer struct {
	sync.Mutex

	config       *config.BalancerConfig
	ipvsMngr     ipvs.Syncer
	iptablesMngr iptables.Syncer
	bgpMngr      bgp.Syncer
	vipMngr      vip.Syncer
	ipam         ipam.Allocator
	metrics      metrics.Collector

	store      store.Store
	state      state.State
	shutdownCh chan bool
}

// NewBalancer initializes a new balancer
//TODO: Graceful shutdown on initialization errors
func NewBalancer(config *config.BalancerConfig) (*Balancer, error) {
	store, err := store.New(config)
	if err != nil {
		return nil, err
	}

	state, err := state.New(store, config)
	if err != nil {
		return nil, err
	}

	ipvsMngr, err := ipvs.New()
	if err != nil {
		return nil, err
	}

	vipMngr, err := vip.New(config)
	if err != nil {
		return nil, err
	}

	iptablesMngr, err := iptables.New(config)
	if err != nil {
		return nil, err
	}

	ipam, err := ipam.New(state, config)
	if err != nil {
		return nil, err
	}

	metrics := metrics.NewMetrics(state, config)

	balancer := &Balancer{
		store:        store,
		state:        state,
		ipvsMngr:     ipvsMngr,
		iptablesMngr: iptablesMngr,
		vipMngr:      vipMngr,
		config:       config,
		ipam:         ipam,
		metrics:      metrics,
	}

	if balancer.isAnycast() {
		bgpMngr, err := bgp.NewBgpService(config)
		if err != nil {
			return nil, err
		}

		balancer.bgpMngr = bgpMngr

		go bgpMngr.Serve()
	}

	/* Cleanup all VIPs on the network interface */
	if err := fusis_net.DelVips(balancer.config.Interfaces.Inbound); err != nil {
		return nil, fmt.Errorf("error cleaning up network vips: %v", err)
	}

	// go balancer.watchLeaderChanges()
	go balancer.watchState()
	// go balancer.watchHealthChecks()

	go metrics.Monitor()

	return balancer, nil
}

func (b *Balancer) getLibLogOutput() io.Writer {
	minLevel := strings.ToUpper(b.config.LogLevel)
	level, _ := log.ParseLevel(minLevel)
	log.SetLevel(level)

	filter := &logutils.LevelFilter{
		Levels:   config.LOG_LEVELS,
		MinLevel: logutils.LogLevel(minLevel),
		Writer:   log.StandardLogger().Out,
	}

	return filter
}

func (b *Balancer) watchState() {
	for {
		select {
		case _ = <-b.state.ChangesCh():
			// TODO: this doesn't need to run all the time, we can implement
			// some kind of throttling in the future waiting for a threashold of
			// messages before applying the messages.
			b.handleStateChange()
		}
	}
}

func (b *Balancer) handleStateChange() error {
	if err := b.ipvsMngr.Sync(b.state); err != nil {
		return err
	}

	if err := b.iptablesMngr.Sync(b.state); err != nil {
		return err
	}

	if b.isAnycast() {
		if err := b.bgpMngr.Sync(b.state); err != nil {
			return err
		}
	} else if !b.IsLeader() {
		return nil
	}

	if err := b.vipMngr.Sync(b.state); err != nil {
		return err
	}

	return nil
}

func (b *Balancer) watchHealthChecks() {
	// for {
	// 	check := <-b.state.HealthCheckCh()
	// 	if err := b.UpdateCheck(check); err != nil {
	// 		log.Error(errors.Wrap(err, "Updating Check failed"))
	// 	}
	// }
}

func (b *Balancer) IsLeader() bool {
	fmt.Println("Implement")
	return true
}

func (b *Balancer) GetLeader() string {
	fmt.Println("Implement")
	return ""
}

func (b *Balancer) watchLeaderChanges() {
	// if b.isAnycast() {
	// 	return
	// }
	//
	// for {
	// 	isLeader := <-b.raft.LeaderCh()
	// 	if isLeader {
	// 		if err := b.vipMngr.Sync(*b.state); err != nil {
	// 			log.Fatal("Could not sync Vips", err)
	// 		}
	//
	// 		if err := b.sendGratuitousARPReply(); err != nil {
	// 			log.Errorf("error sending Gratuitous ARP Reply")
	// 		}
	//
	// 	} else {
	// 		b.flushVips()
	// 	}
	// }
}

func (b *Balancer) sendGratuitousARPReply() error {
	for _, s := range b.GetServices() {
		if err := fusis_net.SendGratuitousARPReply(s.Address, b.config.Interfaces.Inbound); err != nil {
			return err
		}
	}

	return nil
}

func (b *Balancer) flushVips() {
	if err := fusis_net.DelVips(b.config.Interfaces.Inbound); err != nil {
		//TODO: Remove balancer from cluster when error occurs
		log.Error(err)
	}
}

func (b *Balancer) Shutdown() {
}

func (b *Balancer) isAnycast() bool {
	return b.config.ClusterMode == "anycast"
}
