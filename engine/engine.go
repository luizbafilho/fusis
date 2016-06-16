package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider"
)

// Engine ...
type Engine struct {
	sync.Mutex

	Ipvs      *ipvs.Ipvs
	State     ipvs.State
	Provider  provider.Provider
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
	Service     *types.Service
	Destination *types.Destination
}

// New creates a new Engine
func New() (*Engine, error) {
	state := ipvs.NewFusisState()
	ipvsInstance, err := ipvs.New()
	if err != nil {
		return nil, err
	}

	return &Engine{
		CommandCh: make(chan Command),
		State:     state,
		Ipvs:      ipvsInstance,
	}, nil
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

func (e *Engine) applyAddService(svc *types.Service) error {
	if err := e.Ipvs.AddService(ipvs.ToIpvsService(svc)); err != nil {
		return err
	}

	e.State.AddService(svc)

	return nil
}

func (e *Engine) applyDelService(svc *types.Service) error {
	if err := e.Ipvs.DeleteService(ipvs.ToIpvsService(svc)); err != nil {
		return err
	}

	e.State.DeleteService(svc)
	return nil
}

func (e *Engine) applyAddDestination(svc *types.Service, dst *types.Destination) error {
	err := e.Ipvs.AddDestination(*ipvs.ToIpvsService(svc), *ipvs.ToIpvsDestination(dst))
	if err != nil {
		return nil
	}

	e.State.AddDestination(dst)

	return nil
}

func (e *Engine) applyDelDestination(svc *types.Service, dst *types.Destination) error {
	err := e.Ipvs.DeleteDestination(*ipvs.ToIpvsService(svc), *ipvs.ToIpvsDestination(dst))
	if err != nil {
		return nil
	}

	e.State.DeleteDestination(dst)

	return nil
}

type fusisSnapshot struct {
	Services []types.Service
}

func (e *Engine) Snapshot() (raft.FSMSnapshot, error) {
	logrus.Info("Snapshotting Fusis State")
	e.Lock()
	defer e.Unlock()

	services := e.State.GetServices()

	return &fusisSnapshot{services}, nil
}

// Restore stores the key-value store to a previous state.
func (e *Engine) Restore(rc io.ReadCloser) error {
	logrus.Info("Restoring Fusis state")
	var services []types.Service
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
