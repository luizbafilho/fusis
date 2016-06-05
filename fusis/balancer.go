package fusis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/ipvs"
	fusis_net "github.com/luizbafilho/fusis/net"
	_ "github.com/luizbafilho/fusis/provider/none" // to intialize

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
	raftTransport *raft.NetworkTransport
	logger        *logrus.Logger

	engine     *engine.Engine
	shutdownCh chan bool
}

// NewBalancer initializes a new balancer
//TODO: Graceful shutdown on initialization errors
func NewBalancer() (*Balancer, error) {
	engine, err := engine.New()
	if err != nil {
		return nil, err
	}

	balancer := &Balancer{
		eventCh: make(chan serf.Event, 64),
		engine:  engine,
		logger:  logrus.New(),
	}

	if err = balancer.setupRaft(); err != nil {
		log.Fatalf("Setuping Raft", err)
	}

	if err = balancer.setupSerf(); err != nil {
		log.Fatalf("Setuping Serf", err)
	}

	// Flushing all VIPs on the network interface
	if err := fusis_net.DelVips(config.Balancer.Provider.Params["interface"]); err != nil {
		log.Fatalf("Fusis wasn't capable of cleanup network vips. Err: %v", err)
	}

	go balancer.watchLeaderChanges()

	return balancer, nil
}

// Start starts the balancer
func (b *Balancer) setupSerf() error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "balancer"
	conf.Tags["raft-port"] = strconv.Itoa(config.Balancer.RaftPort)

	bindAddr, err := config.Balancer.GetIpByInterface()
	if err != nil {
		return err
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

func (b *Balancer) newStdLogger() *log.Logger {
	return log.New(b.logger.Writer(), "", 0)
}

func (b *Balancer) setupRaft() error {
	// Setup Raft configuration.
	raftConfig := raft.DefaultConfig()
	raftConfig.Logger = b.newStdLogger()

	raftConfig.ShutdownOnRemove = false
	// Check for any existing peers.
	peers, err := readPeersJSON(filepath.Join(config.Balancer.ConfigPath, "peers.json"))
	if err != nil {
		return err
	}

	// Allow the node to entry single-mode, potentially electing itself, if
	// explicitly enabled and there is only 1 node in the cluster already.
	if config.Balancer.Single && len(peers) <= 1 {
		b.logger.Infof("enabling single-node mode")
		raftConfig.EnableSingleNode = true
		raftConfig.DisableBootstrapAfterElect = false
	}

	ip, err := config.Balancer.GetIpByInterface()
	if err != nil {
		return err
	}

	// Setup Raft communication.
	raftAddr := &net.TCPAddr{IP: net.ParseIP(ip), Port: config.Balancer.RaftPort}
	transport, err := raft.NewTCPTransport(raftAddr.String(), raftAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}
	b.raftTransport = transport

	// Create peer storage.
	peerStore := raft.NewJSONPeers(config.Balancer.ConfigPath, transport)
	b.raftPeers = peerStore

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(config.Balancer.ConfigPath, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(config.Balancer.ConfigPath, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}
	b.raftStore = logStore

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(raftConfig, b.engine, logStore, logStore, snapshots, peerStore, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	b.raft = ra

	go b.watchCommands()

	return nil
}

func (b *Balancer) watchCommands() {
	for {
		select {
		case c := <-b.engine.CommandCh:
			switch c.Op {
			case engine.AddServiceOp:
				b.AssignVIP(c.Service)
			case engine.DelServiceOp:
				b.UnassignVIP(c.Service)
			}
		}
	}
}

func (b *Balancer) UnassignVIP(svc *ipvs.Service) {
	if b.isLeader() {
		if err := b.engine.UnassignVIP(svc); err != nil {
			b.logger.Errorf("Unassigning VIP to Service: %#v. Err: %#v", svc, err)
		}
	}
}

func (b *Balancer) AssignVIP(svc *ipvs.Service) {
	if b.isLeader() {
		if err := b.engine.AssignVIP(svc); err != nil {
			b.logger.Errorf("Assigning VIP to Service: %#v. Err: %#v", svc, err)
		}
	}
}

func (b *Balancer) isLeader() bool {
	return b.raft.State() == raft.Leader
}

// JoinPool joins the Fusis Serf cluster
func (b *Balancer) JoinPool() error {
	b.logger.Infof("Balancer: joining: %v ignore: %v", config.Balancer.Join)

	_, err := b.serf.Join([]string{config.Balancer.Join}, true)
	if err != nil {
		b.logger.Errorf("Balancer: error joining: %v", err)
		return err
	}

	return nil
}

func (b *Balancer) watchLeaderChanges() {
	b.logger.Infof("Watching to Leader changes")

	for {
		if <-b.raft.LeaderCh() {
			b.flushVips()
			b.setVips()
		} else {
			b.flushVips()
		}
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
			// case serf.EventQuery:
			// 	query := e.(*serf.Query)
			// 	b.handleQuery(query)
			default:
				b.logger.Warnf("Balancer: unhandled Serf Event: %#v", e)
			}
		}
	}
}

func (b *Balancer) setVips() {
	svcs := b.engine.State.GetServices()

	for _, s := range *svcs {
		err := b.engine.AssignVIP(&s)
		if err != nil {
			//TODO: Remove balancer from cluster when error occurs
			b.logger.Error(err)
		}
	}
}

func (b *Balancer) flushVips() {
	if err := fusis_net.DelVips(config.Balancer.Provider.Params["interface"]); err != nil {
		panic(err)
	}
}

func (b *Balancer) handleMemberJoin(event serf.MemberEvent) {
	b.logger.Infof("handleMemberJoin: %s", event)

	if !b.isLeader() {
		return
	}

	for _, m := range event.Members {
		if isBalancer(m) {
			b.addMemberToPool(m)
		}
	}
}

func (b *Balancer) addMemberToPool(m serf.Member) {
	remoteAddr := fmt.Sprintf("%s:%v", m.Addr.String(), config.Balancer.RaftPort)

	b.logger.Infof("Adding Balancer to Pool", remoteAddr)
	f := b.raft.AddPeer(remoteAddr)
	if f.Error() != nil {
		b.logger.Errorf("node at %s joined failure. err: %s", remoteAddr, f.Error())
	}
}

func isBalancer(m serf.Member) bool {
	return m.Tags["role"] == "balancer"
}

func (b *Balancer) handleMemberLeave(memberEvent serf.MemberEvent) {
	b.logger.Infof("handleMemberLeave: %s", memberEvent)
	for _, m := range memberEvent.Members {
		if isBalancer(m) {
			b.handleBalancerLeave(m)
		} else {
			b.handleAgentLeave(m)
		}
	}
}

func (b *Balancer) handleBalancerLeave(m serf.Member) {
	b.logger.Info("Removing left balancer from raft")
	if !b.isLeader() {
		b.logger.Info("Member is not leader")
		return
	}

	raftPort, err := strconv.Atoi(m.Tags["raft-port"])
	if err != nil {
		b.logger.Errorln("handle balancer leaver failed", err)
	}

	peer := &net.TCPAddr{IP: m.Addr, Port: raftPort}
	b.logger.Infof("Removing %v peer from raft", peer)
	future := b.raft.RemovePeer(peer.String())
	if err := future.Error(); err != nil && err != raft.ErrUnknownPeer {
		b.logger.Errorf("balancer: failed to remove raft peer '%v': %v", peer, err)
	} else if err == nil {
		b.logger.Infof("balancer: removed balancer '%s' as peer", m.Name)
	}
}

func (b *Balancer) Leave() {
	b.logger.Info("balancer: server starting leave")
	// s.left = true

	// Check the number of known peers
	numPeers, err := b.numOtherPeers()
	if err != nil {
		b.logger.Errorf("balancer: failed to check raft peers: %v", err)
		return
	}

	// If we are the current leader, and we have any other peers (cluster has multiple
	// servers), we should do a RemovePeer to safely reduce the quorum size. If we are
	// not the leader, then we should issue our leave intention and wait to be removed
	// for some sane period of time.
	isLeader := b.isLeader()
	// if isLeader && numPeers > 0 {
	// 	future := b.raft.RemovePeer(b.raftTransport.LocalAddr())
	// 	if err := future.Error(); err != nil && err != raft.ErrUnknownPeer {
	// 		b.logger.Errorf("balancer: failed to remove ourself as raft peer: %v", err)
	// 	}
	// }

	// Leave the LAN pool
	if b.serf != nil {
		if err := b.serf.Leave(); err != nil {
			b.logger.Errorf("balancer: failed to leave LAN Serf cluster: %v", err)
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
				b.logger.Errorf("balancer: failed to check raft peers: %v", err)
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
			b.logger.Warnln("balancer: failed to leave raft peer set gracefully, timeout")
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
		b.logger.Errorf("balancer: Error shutting down raft: %s", err)
	}

	if b.raftStore != nil {
		b.raftStore.Close()
	}

	b.raftPeers.SetPeers(nil)
}

func (b *Balancer) handleAgentLeave(m serf.Member) {
	dst, err := b.GetDestination(m.Name)
	if err != nil {
		b.logger.Errorln("handleAgenteLeave failed", err)
		return
	}

	b.DeleteDestination(dst)
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
