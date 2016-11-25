package health

import (
	"sync"
	"time"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/store"
)

type HealthMonitor interface {
	Start()
	FilterHealthy(state state.State) state.State
}

type ServiceID string
type DestinationID string
type Status string

type FusisMonitor struct {
	sync.RWMutex

	store store.Store

	updateCh chan bool

	runningChecks map[string]Check
	desiredChecks map[string]Check

	checkSpecs map[string]types.CheckSpec

	currentDestinations []types.Destination
	currentSpecs        []types.CheckSpec
}

func NewMonitor(store store.Store) HealthMonitor {
	return &FusisMonitor{
		store:               store,
		runningChecks:       make(map[string]Check),
		desiredChecks:       make(map[string]Check),
		currentDestinations: []types.Destination{},
		currentSpecs:        []types.CheckSpec{},
	}
}

func (m *FusisMonitor) getChecksToAdd(running map[string]Check, desired map[string]Check) []Check {
	toAdd := []Check{}

	for _, dCheck := range desired {
		if _, ok := running[dCheck.GetId()]; !ok {
			toAdd = append(toAdd, dCheck)
		}
	}

	return toAdd
}

func (m *FusisMonitor) getChecksToRemove(running map[string]Check, desired map[string]Check) []Check {
	toRemove := []Check{}

	for _, rCheck := range running {
		if _, ok := desired[rCheck.GetId()]; !ok {
			toRemove = append(toRemove, rCheck)
		}
	}

	return toRemove
}

func (m *FusisMonitor) watchChanges() {
	destinationsCh := make(chan []types.Destination)
	m.store.SubscribeDestinations(destinationsCh)

	specsCh := make(chan []types.CheckSpec)
	m.store.SubscribeChecks(specsCh)

	for {
		select {
		case m.currentDestinations = <-destinationsCh:
		case m.currentSpecs = <-specsCh:
		}

		m.handleChanges()
	}

}

func (m *FusisMonitor) handleChanges() {
	currentSpec := types.CheckSpec{
		Interval:  5 * time.Second,
		TCP:       "10.2.2.2",
		ServiceID: "filmes",
	}

	m.currentSpecs = []types.CheckSpec{currentSpec}

	m.populateDesiredChecks()

	toAdd := m.getChecksToAdd(m.runningChecks, m.desiredChecks)
	toRemove := m.getChecksToRemove(m.runningChecks, m.desiredChecks)

	for _, check := range toRemove {
		m.runningChecks[check.GetId()].Stop()
		delete(m.runningChecks, check.GetId())
	}

	for _, check := range toAdd {
		go check.Start()
		m.runningChecks[check.GetId()] = check
	}

}

func (m *FusisMonitor) populateDesiredChecks() {
	// Cleaning desired
	m.desiredChecks = make(map[string]Check)

	for _, dst := range m.currentDestinations {
		for _, spec := range m.currentSpecs {
			if dst.ServiceId == spec.ServiceID {
				check := m.newCheck(spec, dst)
				m.desiredChecks[check.GetId()] = check
			}
		}
	}
}

func (m *FusisMonitor) handleSpecsChanges(specs []types.CheckSpec) {
	for _, spec := range specs {
		m.checkSpecs[spec.ServiceID] = spec
	}
}

func (m *FusisMonitor) newCheck(spec types.CheckSpec, dst types.Destination) Check {
	switch spec.Type {
	default:
		check := CheckTCP{Spec: spec, DestinationID: dst.GetId()}
		check.Init(m.updateCh)
		return &check
	}
}

func (m *FusisMonitor) Start() {
	go m.watchChanges()
}

func (m *FusisMonitor) FilterHealthy(state state.State) state.State {
	return state
}
