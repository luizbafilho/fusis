package ipvs

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/docker/libnetwork/ipvs"
	"github.com/labstack/gommon/log"
	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/types"
)

type Ipvs struct {
	sync.Mutex
	handler *ipvs.Handle
}

type Syncer interface {
	Sync(state state.State) error
}

func loadIpvsModule() error {
	return exec.Command("modprobe", "ip_vs").Run()
}

//New creates a new ipvs struct and flushes the IPVS Table
func New() (*Ipvs, error) {
	handler, err := ipvs.New("")
	if err != nil {
		return nil, fmt.Errorf("[ipvs] Initialisation failed: %v", err)
	}

	if err := handler.Flush(); err != nil {
		return nil, fmt.Errorf("[ipvs] Flushing table failed: %v", err)
	}

	return &Ipvs{
		handler: handler,
	}, nil
}

// Sync syncs all ipvs rules present in state to kernel
func (ipvs *Ipvs) Sync(state state.State) error {
	start := time.Now()
	defer func() {
		log.Debugf("[ipvs] Sync took %v", time.Since(start))
	}()

	ipvs.Lock()
	defer ipvs.Unlock()
	log.Debug("[ipvs] Syncing")

	stateSet := ipvs.getStateServicesSet(state)
	currentSet, err := ipvs.getCurrentServicesSet()
	if err != nil {
		return err
	}

	rulesToAdd := stateSet.Difference(currentSet)
	rulesToRemove := currentSet.Difference(stateSet)

	// Adding services and destinations missing
	for r := range rulesToAdd.Iter() {
		service := r.(types.Service)
		dsts := state.GetDestinations(&service)

		if err := ipvs.addServiceAndDestinations(service, dsts); err != nil {
			return err
		}
		log.Debugf("[ipvs] Added service: %#v with destinations: %#v", service, dsts)
	}

	// Cleaning rules
	for r := range rulesToRemove.Iter() {
		service := r.(types.Service)
		err := ipvs.handler.DelService(ToIpvsService(&service))
		if err != nil {
			return err
		}
		log.Debugf("[ipvs] Removed service: %#v", service)
	}

	// Syncing destination rules
	for _, s := range state.GetServices() {
		if err := ipvs.syncDestinations(state, s); err != nil {
			return err
		}
	}

	return nil
}

func (ipvs *Ipvs) syncDestinations(state state.State, svc types.Service) error {
	stateSet := ipvs.getStateDestinationsSet(state, svc)
	currentSet, err := ipvs.getCurrentDestinationsSet(svc)
	if err != nil {
		return err
	}

	rulesToAdd := stateSet.Difference(currentSet)
	rulesToRemove := currentSet.Difference(stateSet)

	for r := range rulesToAdd.Iter() {
		destination := r.(types.Destination)
		if err := ipvs.handler.NewDestination(ToIpvsService(&svc), ToIpvsDestination(&destination)); err != nil {
			return err
		}
	}

	for r := range rulesToRemove.Iter() {
		destination := r.(types.Destination)
		err := ipvs.handler.DelDestination(ToIpvsService(&svc), ToIpvsDestination(&destination))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipvs *Ipvs) addServiceAndDestinations(svc types.Service, dsts []types.Destination) error {
	ipvsService := ToIpvsService(&svc)
	err := ipvs.handler.NewService(ipvsService)
	if err != nil {
		return err
	}

	for _, d := range dsts {
		err := ipvs.handler.NewDestination(ipvsService, ToIpvsDestination(&d))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipvs *Ipvs) getStateServicesSet(state state.State) mapset.Set {
	stateSet := mapset.NewSet()
	for _, s := range state.GetServices() {
		s.Name = ""
		s.Mode = ""
		stateSet.Add(s)
	}

	return stateSet
}

func (i *Ipvs) getCurrentServicesSet() (mapset.Set, error) {
	svcs, err := i.handler.GetServices()
	if err != nil {
		return nil, err
	}

	currentSet := mapset.NewSet()
	for _, s := range svcs {
		currentSet.Add(FromService(s))
	}

	return currentSet, nil
}

func (ipvs *Ipvs) getStateDestinationsSet(state state.State, svc types.Service) mapset.Set {
	// checks := state.GetChecks()
	stateSet := mapset.NewSet()

	// Filter healthy destinations
	for _, d := range state.GetDestinations(&svc) {
		// if check, ok := checks[d.GetId()]; ok {
		// 	if check.Status == health.BAD {
		// 		continue
		// 	}
		// } else { // no healthcheck found
		// 	continue
		// }

		// Clean up to match services from kernel
		d.Name = ""
		d.ServiceId = ""
		stateSet.Add(d)
	}

	return stateSet
}

func (ipvs *Ipvs) getCurrentDestinationsSet(svc types.Service) (mapset.Set, error) {
	currentSet := mapset.NewSet()
	// ipvsSvc, err := ipvs.handler.ListServices(ToIpvsService(&svc))
	// if err != nil {
	// 	return nil, err
	// }
	dsts, err := ipvs.handler.GetDestinations(ToIpvsService(&svc))
	if err != nil {
		return nil, err
	}

	for _, d := range dsts {
		currentSet.Add(fromDestination(d))
	}

	return currentSet, nil
}

// Flush flushes all services and destinations from the IPVS table.
func (i *Ipvs) Flush() error {
	return i.handler.Flush()
}
