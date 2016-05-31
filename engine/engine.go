package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider"
	"github.com/luizbafilho/fusis/steps"
)

// Engine ...
type Engine struct {
	sync.Mutex

	CommandCh chan Command
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
	return &Engine{
		CommandCh: make(chan Command),
	}
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
		e.CommandCh <- c
	case DelServiceOp:
		if err := e.applyDelService(c.Service); err != nil {
			logrus.Error(err)
			return err
		}
		e.CommandCh <- c
	case AddDestinationOp:
		if err := e.applyAddDestination(c.Service, c.Destination); err != nil {
			logrus.Error(err)
			return err
		}
		e.CommandCh <- c
	case DelDestinationOp:
		if err := e.applyDelDestination(c.Service, c.Destination); err != nil {
			logrus.Error(err)
			return err
		}
		e.CommandCh <- c
	}
	return nil
}

func (e *Engine) applyAddService(svc *ipvs.Service) error {
	seq := steps.NewSequence(
		addServiceStore{svc},
		addServiceIpvs{svc},
	)

	if err := seq.Execute(); err != nil {
		return err
	}

	return nil
}

func (e *Engine) applyDelService(svc *ipvs.Service) error {
	seq := steps.NewSequence(
		deleteServiceIpvs{svc},
		deleteServiceStore{svc},
	)

	if err := seq.Execute(); err != nil {
		return err
	}

	return nil
}

func (e *Engine) applyAddDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
	seq := steps.NewSequence(
		addDestinationStore{svc, dst},
		addDestinationIpvs{svc, dst},
	)

	if err := seq.Execute(); err != nil {
		return err
	}

	return nil
}

func (e *Engine) applyDelDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
	seq := steps.NewSequence(
		deleteDestinationStore{svc, dst},
		deleteDestinationIpvs{svc, dst},
	)

	if err := seq.Execute(); err != nil {
		logrus.Errorln("rolling back applyDelDestination()")
		return err
	}

	return nil
}

func (e *Engine) AssignVIP(svc *ipvs.Service) error {
	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	return prov.AssignVIP(*svc)
}

func (e *Engine) UnassignVIP(svc *ipvs.Service) error {
	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	return prov.UnassignVIP(*svc)
}

type fusisSnapshot struct {
	Services *[]ipvs.Service
}

func (e *Engine) Snapshot() (raft.FSMSnapshot, error) {
	logrus.Info("Snapshotting Fusis State")
	e.Lock()
	defer e.Unlock()

	services, err := ipvs.Store.GetServices()
	if err != nil {
		return nil, err
	}

	return &fusisSnapshot{services}, nil
}

// Restore stores the key-value store to a previous state.
func (e *Engine) Restore(rc io.ReadCloser) error {
	logrus.Info("Restoring Fusis state")
	var services []ipvs.Service
	if err := json.NewDecoder(rc).Decode(&services); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	for _, s := range services {
		if err := e.applyAddService(&s); err != nil {
			return err
		}

		for _, d := range s.Destinations {
			if err := e.applyAddDestination(&s, &d); err != nil {
				return err
			}
		}
	}

	return nil
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
