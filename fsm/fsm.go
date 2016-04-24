package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/steps"
)

// FusisFSM ...
type FusisFSM struct {
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

type Command struct {
	Op          int
	Service     *ipvs.Service
	Destination *ipvs.Destination
}

func New(db *storm.DB) *FusisFSM {
	return &FusisFSM{
		Db: db,
	}
}

// Apply actions to fsm
func (f *FusisFSM) Apply(l *raft.Log) interface{} {
	var c Command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	spew.Dump(c)
	logrus.Infof("Actions received to be aplied to fsm: %v", c)
	switch c.Op {
	case AddServiceOp:
		if err := f.applyAddService(c.Service); err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

func (f *FusisFSM) applyAddService(svc *ipvs.Service) error {
	seq := steps.NewSequence(
		addServiceStore{svc},
		addServiceIpvs{svc},
	)

	return seq.Execute()
}

//Snapshot returns a snapshot of the key-value store.
func (e *FusisFSM) Snapshot() (raft.FSMSnapshot, error) {
	logrus.Info("Chamando snapshot")
	e.Lock()
	defer e.Unlock()

	return &fusisSnapshot{}, nil
}

// Restore stores the key-value store to a previous state.
func (e *FusisFSM) Restore(rc io.ReadCloser) error {
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
