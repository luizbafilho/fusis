package fusis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	fusis_net "github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/provider"
	"github.com/luizbafilho/fusis/state"

	"github.com/hashicorp/logutils"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/hashicorp/serf/serf"
)

const (
	retainSnapshotCount   = 2
	raftTimeout           = 10 * time.Second
	raftRemoveGracePeriod = 5 * time.Second
)

// Balancer represents the Load Balancer
type Balancer struct {
	sync.Mutex
	eventCh chan serf.Event

	serf          *serf.Serf
	raft          *raft.Raft // The consensus mechanism
	raftPeers     raft.PeerStore
	raftStore     *raftboltdb.BoltStore
	raftInmem     *raft.InmemStore
	raftTransport *raft.NetworkTransport
	config        *config.BalancerConfig

	state      *state.State
	provider   provider.Provider
	shutdownCh chan bool
}

// NewBalancer initializes a new balancer
//TODO: Graceful shutdown on initialization errors
func NewBalancer(config *config.BalancerConfig) (*Balancer, error) {
	provider, err := provider.New(config)
	if err != nil {
		return nil, err
	}

	state, err := state.New(config)
	if err != nil {
		return nil, err
	}

	balancer := &Balancer{
		eventCh:  make(chan serf.Event, 64),
		state:    state,
		provider: provider,
		config:   config,
	}

	if err = balancer.setupRaft(); err != nil {
		return nil, fmt.Errorf("error setting up Raft: %v", err)
	}

	if err = balancer.setupSerf(); err != nil {
		return nil, fmt.Errorf("error setting up Serf: %v", err)
	}

	// Flushing all VIPs on the network interface
	if err := fusis_net.DelVips(balancer.config.Provider.Params["interface"]); err != nil {
		return nil, fmt.Errorf("error cleaning up network vips: %v", err)
	}

	go balancer.watchLeaderChanges()

	// Only collect stats if some interval is defined
	if config.Stats.Interval > 0 {
		go balancer.collectStats()
	}

	return balancer, nil
}

// Start starts the balancer
func (b *Balancer) setupSerf() error {
	conf := serf.DefaultConfig()
	conf.Init()

	conf.LogOutput = b.getLibLogOutput()

	conf.Tags["role"] = "balancer"
	conf.Tags["raft-port"] = strconv.Itoa(b.config.Ports["raft"])

	bindAddr, err := b.config.GetIpByInterface()
	if err != nil {
		return err
	}

	conf.MemberlistConfig.BindAddr = bindAddr
	conf.MemberlistConfig.BindPort = b.config.Ports["serf"]

	conf.NodeName = b.config.Name
	conf.EventCh = b.eventCh

	serf, err := serf.Create(conf)
	if err != nil {
		return err
	}

	b.serf = serf

	go b.handleEvents()

	return nil
}

func (b *Balancer) getLibLogOutput() io.Writer {
	logger := log.StandardLogger()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   logger.Out,
	}

	return filter
}

func (b *Balancer) setupRaft() error {
	// Setup Raft configuration.
	raftConfig := raft.DefaultConfig()
	raftConfig.LogOutput = b.getLibLogOutput()

	raftConfig.ShutdownOnRemove = false
	// Check for any existing peers.
	peers, err := readPeersJSON(filepath.Join(b.config.ConfigPath, "peers.json"))
	if err != nil {
		return err
	}

	// Allow the node to entry single-mode, potentially electing itself, if
	// explicitly enabled and there is only 1 node in the cluster already.
	if b.config.Bootstrap && len(peers) <= 1 {
		log.Infof("enabling single-node mode")
		raftConfig.EnableSingleNode = true
		raftConfig.DisableBootstrapAfterElect = false
	}

	ip, err := b.config.GetIpByInterface()
	if err != nil {
		return err
	}

	// Setup Raft communication.
	raftAddr := &net.TCPAddr{IP: net.ParseIP(ip), Port: b.config.Ports["raft"]}
	transport, err := raft.NewTCPTransport(raftAddr.String(), raftAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}
	b.raftTransport = transport

	var log raft.LogStore
	var stable raft.StableStore
	var snap raft.SnapshotStore

	if b.config.DevMode {
		store := raft.NewInmemStore()
		b.raftInmem = store
		stable = store
		log = store
		snap = raft.NewDiscardSnapshotStore()
		b.raftPeers = &raft.StaticPeers{}
	} else {
		// Create peer storage.
		peerStore := raft.NewJSONPeers(b.config.ConfigPath, transport)
		b.raftPeers = peerStore

		var snapshots *raft.FileSnapshotStore
		// Create the snapshot store. This allows the Raft to truncate the log.
		snapshots, err = raft.NewFileSnapshotStore(b.config.ConfigPath, retainSnapshotCount, os.Stderr)
		if err != nil {
			return fmt.Errorf("file snapshot store: %s", err)
		}
		snap = snapshots

		var logStore *raftboltdb.BoltStore
		// Create the log store and stable store.
		logStore, err = raftboltdb.NewBoltStore(filepath.Join(b.config.ConfigPath, "raft.db"))
		if err != nil {
			return fmt.Errorf("new bolt store: %s", err)
		}
		b.raftStore = logStore
		log = logStore
		stable = logStore
	}

	go b.watchState()

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(raftConfig, b.state, log, stable, snap, b.raftPeers, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	b.raft = ra

	return nil
}

func (b *Balancer) watchState() {
	for {
		select {
		case rsp := <-b.state.StateCh:
			// TODO: this doesn't need to run all the time, we can implement
			// some kind of throttling in the future waiting for a threashold of
			// messages before applying the messages.
			rsp <- b.handleStateChange()
		}
	}
}

func (b *Balancer) handleStateChange() error {
	if b.IsLeader() {
		b.provider.SyncVIPs(b.state.Store)
	} else {
		b.Lock()
		defer b.Unlock()
	}
	// return b.state.Ipvs.SyncState(b.state.State)
	return nil
}

func (b *Balancer) IsLeader() bool {
	return b.raft.State() == raft.Leader
}

func (b *Balancer) GetLeader() string {
	return b.raft.Leader()
}

// JoinPool joins the Fusis Serf cluster
func (b *Balancer) JoinPool() error {
	log.Infof("Balancer: joining: %v", b.config.Join)

	_, err := b.serf.Join(b.config.Join, true)
	if err != nil {
		log.Errorf("Balancer: error joining: %v", err)
		return err
	}

	return nil
}

func (b *Balancer) watchLeaderChanges() {
	log.Infof("Watching to Leader changes")

	for {
		isLeader := <-b.raft.LeaderCh()
		b.Lock()
		if isLeader {
			b.flushVips()
			b.setVips()
		} else {
			b.flushVips()
		}
		b.Unlock()
	}
}

func (b *Balancer) handleEvents() {
	for {
		select {
		case e := <-b.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				me := e.(serf.MemberEvent)
				b.handleMemberJoin(me)
			case serf.EventMemberFailed:
				memberEvent := e.(serf.MemberEvent)
				b.handleMemberLeave(memberEvent)
			case serf.EventMemberLeave:
				memberEvent := e.(serf.MemberEvent)
				b.handleMemberLeave(memberEvent)
			default:
				log.Warnf("Balancer: unhandled Serf Event: %#v", e)
			}
		}
	}
}

func (b *Balancer) setVips() {
	err := b.provider.SyncVIPs(b.state.Store)
	if err != nil {
		//TODO: Remove balancer from cluster when error occurs
		log.Error(err)
	}
}

func (b *Balancer) flushVips() {
	if err := fusis_net.DelVips(b.config.Provider.Params["interface"]); err != nil {
		//TODO: Remove balancer from cluster when error occurs
		log.Error(err)
	}
}

func (b *Balancer) handleMemberJoin(event serf.MemberEvent) {
	log.Infof("handleMemberJoin: %s", event)

	if !b.IsLeader() {
		return
	}

	for _, m := range event.Members {
		if isBalancer(m) {
			b.addMemberToPool(m)
		} else {
			b.handleAgentJoin(m)
		}
	}
}

func (b *Balancer) handleAgentJoin(m serf.Member) {
	var dst *types.Destination
	if err := json.Unmarshal([]byte(m.Tags["info"]), &dst); err != nil {
		log.Error("Unable to Unmarshal new destination info", err)
	}

	srv, err := b.GetService(dst.ServiceId)
	if err != nil {
		log.Error("handleAgentJoin: Unable to find service", err)
	}

	if err := b.AddDestination(srv, dst); err != nil {
		log.WithFields(log.Fields{
			"err":         err,
			"destination": dst,
		}).Error("handleAgentJoin: Unable to add new destination")
	}
}

func (b *Balancer) addMemberToPool(m serf.Member) {
	remoteAddr := fmt.Sprintf("%s:%v", m.Addr.String(), m.Tags["raft-port"])

	log.Infof("Adding Balancer to Pool", remoteAddr)
	f := b.raft.AddPeer(remoteAddr)
	if f.Error() != nil {
		log.Errorf("node at %s joined failure. err: %s", remoteAddr, f.Error())
	}
}

func isBalancer(m serf.Member) bool {
	return m.Tags["role"] == "balancer"
}

func (b *Balancer) handleMemberLeave(memberEvent serf.MemberEvent) {
	log.Infof("handleMemberLeave: %s", memberEvent)
	for _, m := range memberEvent.Members {
		if isBalancer(m) {
			b.handleBalancerLeave(m)
		} else {
			b.handleAgentLeave(m)
		}
	}
}

func (b *Balancer) handleBalancerLeave(m serf.Member) {
	log.Info("Removing left balancer from raft")
	if !b.IsLeader() {
		log.Info("Member is not leader")
		return
	}

	raftPort, err := strconv.Atoi(m.Tags["raft-port"])
	if err != nil {
		log.Errorln("handle balancer leaver failed", err)
	}

	peer := &net.TCPAddr{IP: m.Addr, Port: raftPort}
	log.Infof("Removing %v peer from raft", peer)

	future := b.raft.RemovePeer(peer.String())
	if err := future.Error(); err != nil && err != raft.ErrUnknownPeer {
		log.Errorf("balancer: failed to remove raft peer '%v': %v", peer, err)
	} else if err == nil {
		log.Infof("balancer: removed balancer '%s' as peer", m.Name)
	}
}

func (b *Balancer) Leave() {
	log.Info("balancer: server starting leave")
	// s.left = true

	// Check the number of known peers
	numPeers, err := b.numOtherPeers()
	if err != nil {
		log.Errorf("balancer: failed to check raft peers: %v", err)
		return
	}

	// If we are the current leader, and we have any other peers (cluster has multiple
	// servers), we should do a RemovePeer to safely reduce the quorum size. If we are
	// not the leader, then we should issue our leave intention and wait to be removed
	// for some sane period of time.
	isLeader := b.IsLeader()
	// if isLeader && numPeers > 0 {
	// 	future := b.raft.RemovePeer(b.raftTransport.LocalAddr())
	// 	if err := future.Error(); err != nil && err != raft.ErrUnknownPeer {
	//		log.Errorf("balancer: failed to remove ourself as raft peer: %v", err)
	// 	}
	// }

	// Leave the LAN pool
	if b.serf != nil {
		if err := b.serf.Leave(); err != nil {
			log.Errorf("balancer: failed to leave LAN Serf cluster: %v", err)
		}
	}

	// If we were not leader, wait to be safely removed from the cluster.
	// We must wait to allow the raft replication to take place, otherwise
	// an immediate shutdown could cause a loss of quorum.
	if !isLeader {
		limit := time.Now().Add(raftRemoveGracePeriod)
		for numPeers > 0 && time.Now().Before(limit) {
			// Update the number of peers
			numPeers, err = b.numOtherPeers()
			if err != nil {
				log.Errorf("balancer: failed to check raft peers: %v", err)
				break
			}

			// Avoid the sleep if we are done
			if numPeers == 0 {
				break
			}

			// Sleep a while and check again
			time.Sleep(50 * time.Millisecond)
		}
		if numPeers != 0 {
			log.Warnln("balancer: failed to leave raft peer set gracefully, timeout")
		}
	}
}

// numOtherPeers is used to check on the number of known peers
// excluding the local node
func (b *Balancer) numOtherPeers() (int, error) {
	peers, err := b.raftPeers.Peers()
	if err != nil {
		return 0, err
	}
	otherPeers := raft.ExcludePeer(peers, b.raftTransport.LocalAddr())
	return len(otherPeers), nil
}

func (b *Balancer) Shutdown() {
	b.Leave()
	b.serf.Shutdown()

	future := b.raft.Shutdown()
	if err := future.Error(); err != nil {
		log.Errorf("balancer: Error shutting down raft: %s", err)
	}

	if b.raftStore != nil {
		b.raftStore.Close()
	}

	b.raftPeers.SetPeers(nil)
}

func (b *Balancer) handleAgentLeave(m serf.Member) {
	dst, err := b.GetDestination(m.Name)
	if err != nil {
		log.Errorln("handleAgenteLeave failed", err)
		return
	}

	b.DeleteDestination(dst)
}

func (b *Balancer) collectStats() {

	// interval := b.config.Stats.Interval
	//
	// if interval > 0 {
	// 	ticker := time.NewTicker(time.Second * time.Duration(interval))
	// 	for tick := range ticker.C {
	// 		// b.state.CollectStats(tick)
	// 	}
	// }
}

func readPeersJSON(path string) ([]string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(b) == 0 {
		return nil, nil
	}

	var peers []string
	dec := json.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(&peers); err != nil {
		return nil, err
	}

	return peers, nil
}
