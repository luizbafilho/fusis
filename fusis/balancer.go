package fusis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/ipvs"
	fusis_net "github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/provider"
	_ "github.com/luizbafilho/fusis/provider/none" // to intialize

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/hashicorp/serf/serf"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

// Balancer represents the Load Balancer
type Balancer struct {
	eventCh chan serf.Event

	serf *serf.Serf
	raft *raft.Raft // The consensus mechanism

	engine   *engine.Engine
	provider provider.Provider
}

// NewBalancer initializes a new balancer
//TODO: Graceful shutdown on initialization errors
func NewBalancer() (*Balancer, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	balancer := &Balancer{
		eventCh:  make(chan serf.Event, 64),
		provider: prov,
	}

	if err = balancer.setupRaft(); err != nil {
		log.Errorln("Setuping Raft")
		panic(err)
	}

	if err = balancer.setupSerf(); err != nil {
		log.Errorln("Setuping Serf")
		panic(err)
	}

	//Initializes ipvs store (boltdb) and ipvs module
	if err := ipvs.Init(); err != nil {
		panic(err)
	}

	// Flushing all VIPs on the network interface
	if err := fusis_net.DelVips(config.Balancer.Provider.Params["interface"]); err != nil {
		panic(err)
	}

	return balancer, nil
}

// Start starts the balancer
func (b *Balancer) setupSerf() error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "balancer"

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

func (b *Balancer) setupRaft() error {
	// Setup Raft configuration.
	raftConfig := raft.DefaultConfig()

	// Check for any existing peers.
	peers, err := readPeersJSON(filepath.Join(config.Balancer.ConfigPath, "peers.json"))
	if err != nil {
		return err
	}

	// Allow the node to entry single-mode, potentially electing itself, if
	// explicitly enabled and there is only 1 node in the cluster already.
	if config.Balancer.Single && len(peers) <= 1 {
		log.Info("enabling single-node mode")
		raftConfig.EnableSingleNode = true
		raftConfig.DisableBootstrapAfterElect = false
	}

	// Setup Raft communication.
	raftAddr := &net.TCPAddr{Port: config.Balancer.RaftPort}
	transport, err := raft.NewTCPTransport(raftAddr.String(), raftAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create peer storage.
	peerStore := raft.NewJSONPeers(config.Balancer.ConfigPath, transport)

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

	engine := engine.New()

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(raftConfig, engine, logStore, logStore, snapshots, peerStore, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	b.raft = ra
	engine.Raft = ra
	return nil
}

// JoinPool joins the Fusis Serf cluster
func (b *Balancer) JoinPool() error {
	log.Infof("Balancer: joining: %v ignore: %v", config.Balancer.Join)

	_, err := b.serf.Join([]string{config.Balancer.Join}, true)
	if err != nil {
		log.Errorf("Balancer: error joining: %v", err)
		return err
	}

	return nil
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
			case serf.EventQuery:
				query := e.(*serf.Query)
				b.handleQuery(query)
			default:
				log.Warnf("Balancer: unhandled Serf Event: %#v", e)
			}
		}
	}
}

func (b *Balancer) handleMemberJoin(event serf.MemberEvent) {
	log.Infof("handleMemberJoin: %s", event)

	if b.raft.State() != raft.Leader {
		return
	}

	for _, m := range event.Members {
		if memberIsBalancer(m) {
			b.addMemberToPool(m)
		}
	}

}

func (b *Balancer) addMemberToPool(m serf.Member) {
	remoteAddr := fmt.Sprintf("%s:%v", m.Addr.String(), config.Balancer.RaftPort)

	log.Infof("addMemberToPool, %#v", remoteAddr)
	f := b.raft.AddPeer(remoteAddr)
	if f.Error() != nil {
		log.Errorf("node at %s joined failure. err: %s", remoteAddr, f.Error())
	}
}

func memberIsBalancer(m serf.Member) bool {
	return m.Tags["role"] == "balancer"
}

func (b *Balancer) handleMemberLeave(member serf.MemberEvent) {
	for _, m := range member.Members {
		DeleteDestination(m.Name)
	}
}

func (b *Balancer) handleQuery(query *serf.Query) {
	payload := query.Payload

	var dst ipvs.Destination
	err := json.Unmarshal(payload, &dst)
	if err != nil {
		log.Errorf("Balancer: Unable to Unmarshal: %s", payload)
	}

	svc, err := GetService(dst.ServiceId)
	if err != nil {
		panic(err)
	}

	err = AddDestination(svc, &dst)
	if err != nil {
		panic(err)
	}
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
