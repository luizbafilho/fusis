package fusis

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/bgp"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/health"
	"github.com/luizbafilho/fusis/ipam"
	"github.com/luizbafilho/fusis/iptables"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/metrics"
	fusis_net "github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/store"
	"github.com/luizbafilho/fusis/types"
	"github.com/luizbafilho/fusis/vip"
	"github.com/luizbafilho/leadership"
	"github.com/pkg/errors"
)

type Balancer interface {
	GetServices() []types.Service
	AddService(*types.Service) error
	GetService(string) (*types.Service, error)
	DeleteService(string) error

	AddDestination(*types.Service, *types.Destination) error
	GetDestination(string) (*types.Destination, error)
	GetDestinations(svc *types.Service) []types.Destination
	DeleteDestination(*types.Destination) error

	AddCheck(check types.CheckSpec) error
	DeleteCheck(check types.CheckSpec) error

	IsLeader() bool

	Shutdown()
}

// Balancer represents the Load Balancer
type FusisBalancer struct {
	sync.RWMutex

	config        *config.BalancerConfig
	ipvsMngr      ipvs.Syncer
	iptablesMngr  iptables.Syncer
	bgpMngr       bgp.Syncer
	vipMngr       vip.Syncer
	ipam          ipam.Allocator
	metrics       metrics.Collector
	healthMonitor health.HealthMonitor

	store     store.Store
	state     state.State
	candidate *leadership.Candidate

	changesCh  chan bool
	shutdownCh chan bool
}

// NewBalancer initializes a new balancer
//TODO: Graceful shutdown on initialization errors
func NewBalancer(config *config.BalancerConfig) (Balancer, error) {
	store, err := store.New(config)
	if err != nil {
		return nil, err
	}

	state, err := state.New()
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

	changesCh := make(chan bool)
	balancer := &FusisBalancer{
		changesCh:    changesCh,
		store:        store,
		state:        state,
		ipvsMngr:     ipvsMngr,
		iptablesMngr: iptablesMngr,
		vipMngr:      vipMngr,
		config:       config,
		ipam:         ipam,
		metrics:      metrics,
	}

	if balancer.config.EnableHealthChecks {
		monitor := health.NewMonitor(store, changesCh)
		go monitor.Start()
		balancer.healthMonitor = monitor
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
		return nil, fmt.Errorf("Error cleaning up network vips: %v", err)
	}

	go balancer.watchLeaderChanges()
	go balancer.watchStore()
	go balancer.watchState()

	go metrics.Monitor()

	balancer.loadInitialState()

	return balancer, nil
}

func (b *FusisBalancer) loadInitialState() error {
	log.Info("[balancer] Loading initial state from Store")
	initSvcs, err := b.store.GetServices()
	if err != nil {
		return errors.Wrap(err, "[balancer] Fetching initial services failed")
	}

	initDsts, err := b.store.GetDestinations()
	if err != nil {
		return errors.Wrap(err, "[balancer] Fetching initial destinations failed")
	}

	b.state.UpdateServices(initSvcs)
	b.state.UpdateDestinations(initDsts)

	if err, module := b.handleStateChange(); err != nil {
		log.Errorf("[%s] Error syncing initial state: %s", module, err)
		return err
	}

	return nil
}

func (b *FusisBalancer) watchStore() {
	updateSvcsCh := make(chan []types.Service)
	b.store.SubscribeServices(updateSvcsCh)

	updateDstsCh := make(chan []types.Destination)
	b.store.SubscribeDestinations(updateDstsCh)

	for {
		select {
		case svcs := <-updateSvcsCh:
			b.state.UpdateServices(svcs)
		case dsts := <-updateDstsCh:
			b.state.UpdateDestinations(dsts)
		}

		b.changesCh <- true
	}
}

func (b *FusisBalancer) watchState() {
	for {
		<-b.changesCh
		// TODO: this doesn't need to run all the time, we can implement
		// some kind of throttling in the future waiting for a threashold of
		// messages before applying the messages.
		if err, module := b.handleStateChange(); err != nil {
			log.Errorf("[%s] Error handling state change: %s", module, err)
		}
	}
}

func (b *FusisBalancer) handleStateChange() (error, string) {
	state := b.state

	if b.config.EnableHealthChecks {
		state = b.healthMonitor.FilterHealthy(b.state)
	}

	if err := b.ipvsMngr.Sync(state); err != nil {
		return err, "ipvs"
	}

	if err := b.iptablesMngr.Sync(state); err != nil {
		return err, "iptables"
	}

	if b.isAnycast() {
		if err := b.bgpMngr.Sync(state); err != nil {
			return err, "bgp"
		}
	} else if !b.IsLeader() {
		return nil, ""
	}

	if err := b.vipMngr.Sync(state); err != nil {
		return err, "vip"
	}

	return nil, ""
}

func (b *FusisBalancer) IsLeader() bool {
	b.RLock()
	if b.candidate == nil {
		return false
	}
	b.RUnlock()

	return b.candidate.IsLeader()
}

func (b *FusisBalancer) watchLeaderChanges() {
	candidate := leadership.NewCandidate(b.store.GetKV(), b.config.StorePrefix+"/leader", b.config.Name, 20*time.Second)
	b.Lock()
	b.candidate = candidate
	b.Unlock()

	electedCh, _ := b.candidate.RunForElection()
	if b.isAnycast() {
		return
	}

	for isElected := range electedCh {
		// This loop will run every time there is a change in our leadership
		// status.

		if isElected {
			log.Println("I won the election! I'm now the leader")
			if err := b.vipMngr.Sync(b.state); err != nil {
				log.Fatal("Could not sync Vips", err)
			}

			if err := b.sendGratuitousARPReply(); err != nil {
				log.Errorf(errors.Wrap(err, "Error sending Gratuitous ARP Reply").Error())
			}
		} else {
			log.Println("Lost the election, let's try another time")
			b.flushVips()
		}
	}
}

func (b *FusisBalancer) sendGratuitousARPReply() error {
	for _, s := range b.GetServices() {
		if err := fusis_net.SendGratuitousARPReply(s.Address, b.config.Interfaces.Inbound); err != nil {
			return err
		}
	}

	return nil
}

// Utility method to cleanup state for tests
func (b *FusisBalancer) cleanup() error {
	for _, svc := range b.GetServices() {
		b.state.DeleteService(&svc)

		for _, dst := range b.GetDestinations(&svc) {
			b.state.DeleteDestination(&dst)
		}
	}

	kv := b.store.GetKV()
	kv.DeleteTree(b.config.StorePrefix)

	return nil
}

func (b *FusisBalancer) flushVips() {
	if err := fusis_net.DelVips(b.config.Interfaces.Inbound); err != nil {
		//TODO: Remove balancer from cluster when error occurs
		log.Error(err)
	}
}

func (b *FusisBalancer) Shutdown() {
}

func (b *FusisBalancer) isAnycast() bool {
	return b.config.ClusterMode == "anycast"
}
