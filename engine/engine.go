package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/steps"
)

// Engine ...
type Engine struct {
	Db *storm.DB

	sync.Mutex
}

// Represents possible actions on engine
const (
	AddServiceOp = iota
	DelServiceOp

	AddDestinationOp
	DelDestinationOp
)

// Command represents a command in raft log
type Command struct {
	Op          int
	Service     *ipvs.Service
	Destination *ipvs.Destination
}

// New creates a new Engine
func New() *Engine {
	return &Engine{}
}

// Apply actions to fsm
func (e *Engine) Apply(l *raft.Log) interface{} {
	var c Command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	logrus.Infof("Actions received to be aplied to fsm: %v", c)
	switch c.Op {
	case AddServiceOp:
		if err := e.applyAddService(c.Service); err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func (e *Engine) applyAddService(svc *ipvs.Service) error {
	seq := steps.NewSequence(
		addServiceStore{svc},
		addServiceIpvs{svc},
		// setVip{svc},
	)

	return seq.Execute()
}

//Snapshot returns a snapshot of the key-value store.
func (e *Engine) Snapshot() (raft.FSMSnapshot, error) {
	logrus.Info("Chamando snapshot")
	e.Lock()
	defer e.Unlock()

	return &fusisSnapshot{}, nil
}

// Restore stores the key-value store to a previous state.
func (e *Engine) Restore(rc io.ReadCloser) error {
	logrus.Info("Chamando restore")
	return nil
}

type fusisSnapshot struct {
}

func (f *fusisSnapshot) Persist(sink raft.SnapshotSink) error {
	logrus.Info("Chamando persist")
	return nil
}

func (f *fusisSnapshot) Release() {
	logrus.Info("Chamando release")
}
