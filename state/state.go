package state

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/health"
)

//go:generate stringer -type=CommandOp

// State...
type State struct {
	sync.Mutex

	services     map[string]types.Service
	destinations map[string]types.Destination
	checks       map[string]*health.Check

	changesCh    chan chan error
	healhCheckCh chan health.Check
}

// Represents possible actions on engine
const (
	AddServiceOp CommandOp = iota
	DelServiceOp
	AddDestinationOp
	DelDestinationOp
	AddCheckOp
	DelCheckOp
	UpdateCheckOp
)

type CommandOp int

// Command represents a command in raft log
type Command struct {
	Op          CommandOp
	Service     *types.Service
	Destination *types.Destination
	Check       *health.Check
	Response    chan interface{} `json:"-"`
}

func (c Command) String() string {
	return fmt.Sprintf("%v: Service: %#v Destination: %#v", c.Op, c.Service, c.Destination)
}

// New creates a new Engine
func New(config *config.BalancerConfig) (*State, error) {
	return &State{
		services:     make(map[string]types.Service),
		destinations: make(map[string]types.Destination),
		checks:       make(map[string]*health.Check),
		changesCh:    make(chan chan error),
		healhCheckCh: make(chan health.Check),
	}, nil
}

func (s *State) ChangesCh() chan chan error {
	return s.changesCh
}

func (s *State) HealthCheckCh() chan health.Check {
	return s.healhCheckCh
}

// Apply actions to fsm
func (s *State) Apply(l *raft.Log) interface{} {
	var c Command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	logrus.Infof("fusis: Action received to be aplied to fsm: %v", c)
	switch c.Op {
	case AddServiceOp:
		s.AddService(c.Service)
	case DelServiceOp:
		s.DeleteService(c.Service)
	case AddDestinationOp:
		s.AddDestination(c.Destination)
	case DelDestinationOp:
		s.DeleteDestination(c.Destination)
	case AddCheckOp:
		s.AddCheck(c.Destination)
	case DelCheckOp:
		s.DeleteCheck(c.Destination)
	case UpdateCheckOp:
		s.UpdateCheck(c.Check)
	}
	rsp := make(chan error)
	s.changesCh <- rsp
	return <-rsp
}

type fusisSnapshot struct {
	Services []types.Service
}

func (s *State) Snapshot() (raft.FSMSnapshot, error) {
	logrus.Info("Snapshotting Fusis State")
	s.Lock()
	defer s.Unlock()

	services := s.GetServices()

	return &fusisSnapshot{services}, nil
}

// Restore stores the key-value store to a previous state.
func (s *State) Restore(rc io.ReadCloser) error {
	logrus.Info("Restoring Fusis state")
	var services []types.Service
	if err := json.NewDecoder(rc).Decode(&services); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	for _, srv := range services {
		s.AddService(&srv)
		//TODO: add destination
		// for _, d := range srv.Destinations {
		// 	s.AddDestination(&d)
		// }
	}
	rsp := make(chan error)
	s.changesCh <- rsp
	return <-rsp
}

func (f *fusisSnapshot) Persist(sink raft.SnapshotSink) error {
	logrus.Infoln("Persisting Fusis state")
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.Services)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		if err := sink.Close(); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (f *fusisSnapshot) Release() {
	logrus.Info("Calling release")
}
