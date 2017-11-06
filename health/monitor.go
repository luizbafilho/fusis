package health

import (
	"fmt"
	"sync"

	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/store"
	"github.com/luizbafilho/fusis/types"
	"github.com/sirupsen/logrus"
)

type HealthMonitor interface {
	Start(changesCh chan bool)
	UpdateChecks(s state.State)
	FilterHealthy(state state.State) state.State
}

type ServiceID string
type DestinationID string
type Status string

type FusisMonitor struct {
	sync.RWMutex

	store store.Store

	changesCh chan bool

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

func (m *FusisMonitor) UpdateChecks(s state.State) {
	m.Lock()
	m.currentDestinations = s.GetDestinations(nil)
	m.currentSpecs = s.GetChecks()
	m.Unlock()

	m.handleStateChange()
}

func (m *FusisMonitor) handleStateChange() {
	m.populateDesiredChecks()

	toAdd := m.getChecksToAdd(m.runningChecks, m.desiredChecks)
	toRemove := m.getChecksToRemove(m.runningChecks, m.desiredChecks)

	for _, check := range toRemove {
		m.Lock()
		m.runningChecks[check.GetId()].Stop()
		delete(m.runningChecks, check.GetId())
		m.Unlock()
	}

	for _, check := range toAdd {
		go check.Start()
		m.Lock()
		m.runningChecks[check.GetId()] = check
		m.Unlock()
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
		check := CheckTCP{Spec: spec, DestinationID: dst.GetId(), Status: BAD}
		check.Init(m.changesCh, dst)
		return &check
	}
}

func (m *FusisMonitor) Start(changesCh chan bool) {
	m.changesCh = changesCh
}

func (m *FusisMonitor) FilterHealthy(s state.State) state.State {

	for _, svc := range s.GetServices() {
		for _, dst := range s.GetDestinations(&svc) {
			checkId := fmt.Sprintf("%s:%s", svc.GetId(), dst.GetId())
			m.RLock()
			if check, ok := m.runningChecks[checkId]; ok {
				if check.GetStatus() == BAD {
					logrus.Debugf("[health-monitor] filtering %#v. Check %s", dst, check.GetId())
					s.DeleteDestination(dst)
				}
			}
			m.RUnlock()
		}
	}

	return s
}
