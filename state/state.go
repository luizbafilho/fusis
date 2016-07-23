package state

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/hashicorp/raft"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
)

//go:generate stringer -type=CommandOp

// State...
type State struct {
	sync.Mutex

	Store   Store
	StateCh chan chan error
}

// Represents possible actions on engine
const (
	AddServiceOp CommandOp = iota
	DelServiceOp
	AddDestinationOp
	DelDestinationOp
)

type CommandOp int

// Command represents a command in raft log
type Command struct {
	Op          CommandOp
	Service     *types.Service
	Destination *types.Destination
	Response    chan interface{} `json:"-"`
}

func (c Command) String() string {
	return fmt.Sprintf("%v: Service: %#v Destination: %#v", c.Op, c.Service, c.Destination)
}

// New creates a new Engine
func New(config *config.BalancerConfig) (*State, error) {
	state := NewFusisState()

	// statsLogger := NewStatsLogger(config)

	return &State{
		StateCh: make(chan chan error),
		Store:   state,
		// StatsLogger: statsLogger,
	}, nil
}

func NewStatsLogger(config *config.BalancerConfig) *logrus.Logger {
	logger := logrus.New()

	if config.Stats.Type == "" {
		return nil
	}

	switch config.Stats.Type {
	case "logstash":
		addLogstashLoggerHook(logger, config)
	case "syslog":
		addSyslogLoggerHook(logger, config)
	default:
		log.Fatal("Unknown stats logger. Please configure properly logstash or syslog.")
	}

	return logger
}

func addSyslogLoggerHook(logger *logrus.Logger, config *config.BalancerConfig) {

	protocol := config.Stats.Params["protocol"]
	address := config.Stats.Params["address"]

	hook, err := logrus_syslog.NewSyslogHook(protocol, address, syslog.LOG_INFO, "")
	if err != nil {
		log.Fatalf("Unable to connect to local syslog daemon. Err: %v", err)
	}

	logger.Hooks.Add(hook)
}

func addLogstashLoggerHook(logger *logrus.Logger, config *config.BalancerConfig) {
	url := fmt.Sprintf("%s:%v", config.Stats.Params["host"], config.Stats.Params["port"])
	hook, err := logrus_logstash.NewHook(config.Stats.Params["protocol"], url, "Fusis")
	if err != nil {
		log.Fatalf("unable to connect to logstash. Err: %v", err)
	}

	logger.Hooks.Add(hook)
}

// Apply actions to fsm
func (e *State) Apply(l *raft.Log) interface{} {
	var c Command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	logrus.Infof("fusis: Action received to be aplied to fsm: %v", c)
	switch c.Op {
	case AddServiceOp:
		e.Store.AddService(c.Service)
	case DelServiceOp:
		e.Store.DeleteService(c.Service)
	case AddDestinationOp:
		e.Store.AddDestination(c.Destination)
	case DelDestinationOp:
		e.Store.DeleteDestination(c.Destination)
	}
	rsp := make(chan error)
	e.StateCh <- rsp
	return <-rsp
}

type fusisSnapshot struct {
	Services []types.Service
}

func (e *State) Snapshot() (raft.FSMSnapshot, error) {
	logrus.Info("Snapshotting Fusis State")
	e.Lock()
	defer e.Unlock()

	services := e.Store.GetServices()

	return &fusisSnapshot{services}, nil
}

// Restore stores the key-value store to a previous state.
func (e *State) Restore(rc io.ReadCloser) error {
	logrus.Info("Restoring Fusis state")
	var services []types.Service
	if err := json.NewDecoder(rc).Decode(&services); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	for _, s := range services {
		e.Store.AddService(&s)
		for _, d := range s.Destinations {
			e.Store.AddDestination(&d)
		}
	}
	rsp := make(chan error)
	e.StateCh <- rsp
	return <-rsp
}

// func (e *State) CollectStats(tick time.Time) {
// 	e.StatsLogger.Info("logging stats")
// 	for _, s := range e.Store.GetServices() {
// 		srv := e.syncService(&s)
//
// 		hosts := []string{}
// 		for _, dst := range srv.Destinations {
// 			hosts = append(hosts, dst.Host)
// 		}
//
// 		e.StatsLogger.WithFields(logrus.Fields{
// 			"time":     tick,
// 			"service":  s.Name,
// 			"Protocol": s.Protocol,
// 			"Port":     s.Port,
// 			"hosts":    strings.Join(hosts, ","),
// 			"client":   "fusis",
// 		}).Info("Fusis router stats")
// 	}
// }

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

// func (e *State) syncService(svc *types.Service) types.Service {
// 	service, err := gipvs.GetService(ipvs.ToIpvsService(svc))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return ipvs.FromService(service)
// }
