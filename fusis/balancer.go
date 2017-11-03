package fusis

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/luizbafilho/fusis/bgp"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/election"
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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Balancer interface {
	GetServices() ([]types.Service, error)
	AddService(*types.Service) error
	GetService(name string) (*types.Service, error)
	DeleteService(string) error

	AddDestination(*types.Service, *types.Destination) error
	// GetDestination(name string) (*types.Destination, error)
	GetDestinations(svc *types.Service) ([]types.Destination, error)
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

	store    store.Store
	state    state.State
	election *election.Election

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

	el, err := election.New(config)
	if err != nil {
		return nil, err
	}

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
		election:     el,
		metrics:      metrics,
	}

	// if balancer.config.EnableHealthChecks {
	// 	monitor := health.NewMonitor(store, changesCh)
	// 	go monitor.Start()
	// 	balancer.healthMonitor = monitor
	// }

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
	// go balancer.watchState()
	go balancer.reconcile()

	go metrics.Monitor()

	if err := balancer.loadState(); err != nil {
		return nil, errors.Wrap(err, "failed to load initial state")
	}

	return balancer, nil
}

func (b *FusisBalancer) reconcile() {
	ticker := time.NewTicker(30 * time.Second)
	for _ = range ticker.C {
		if err := b.loadState(); err != nil {
			log.Errorf("failed to load state in reconcile loop: %s", err)
		}
	}
}

func (b *FusisBalancer) loadState() error {
	s, err := b.store.GetState()
	if err != nil {
		return errors.Wrap(err, "[balancer] get initial state failed")
	}

	if err := b.handleStateChange(s); err != nil {
		log.Errorf("[balancer] error handling initial state: %s", err)
		return err
	}

	return nil
}

func (b *FusisBalancer) watchStore() {
	stateCh := make(chan state.State)
	b.store.AddWatcher(stateCh)

	go b.store.Watch()

	for s := range stateCh {
		if err := b.handleStateChange(s); err != nil {
			log.Errorf("[balancer] Error handling state change: %s", err)
		}
	}
}

func (b *FusisBalancer) handleStateChange(s state.State) error {
	// b.Lock()
	// b.state = s.Copy()
	// b.Unlock()

	start := time.Now()
	defer func() {
		log.Debugf("handleStateChange() took %v", time.Since(start))
	}()

	// if b.config.EnableHealthChecks {
	// 	s = b.healthMonitor.FilterHealthy(s)
	// }

	if err := b.ipvsMngr.Sync(s); err != nil {
		return err
	}

	if err := b.iptablesMngr.Sync(s); err != nil {
		return err
	}

	if b.isAnycast() {
		if err := b.bgpMngr.Sync(s); err != nil {
			return err
		}
	} else if !b.IsLeader() {
		return nil
	}

	if err := b.vipMngr.Sync(s); err != nil {
		return err
	}

	return nil
}

func (b *FusisBalancer) IsLeader() bool {
	return b.election.IsLeader()
}

func (b *FusisBalancer) watchLeaderChanges() {
	// No need to elect a leader when using anycast mode
	if b.isAnycast() {
		return
	}

	log.Debug("[election] running for election")
	for elected := range b.election.Run() {
		// This loop will run every time there is a change in the leadership status.
		if elected {
			log.Info("[election] defined leader")
			if err := b.loadState(); err != nil {
				log.Errorf("failed to load state in reconcile loop: %s", err)
			}

			if err := b.sendGratuitousARPReply(); err != nil {
				log.Errorf(errors.Wrap(err, "error sending gratuitous ARP reply").Error())
			}
		} else {
			log.Info("lost leadership. Flushing VIPs")
			b.flushVips()
		}
	}
}

func (b *FusisBalancer) sendGratuitousARPReply() error {
	svcs, err := b.GetServices()
	if err != nil {
		return err
	}

	for _, s := range svcs {
		if err := fusis_net.SendGratuitousARPReply(s.Address, b.config.Interfaces.Inbound); err != nil {
			return err
		}
	}

	return nil
}

// Utility method to cleanup state for tests
func (b *FusisBalancer) cleanup() error {
	b.flushVips()

	if out, err := exec.Command("ipvsadm", "--clear").CombinedOutput(); err != nil {
		log.Fatal(fmt.Errorf("Running ipvsadm --clear failed with message: `%s`, error: %v", strings.TrimSpace(string(out)), err))
	}

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
