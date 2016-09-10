package ipvs

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/deckarep/golang-set"
	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/state"
)

type Ipvs struct {
	sync.Mutex
}

type Syncer interface {
	Sync(state state.State) error
}

//New creates a new ipvs struct and flushes the IPVS Table
func New() (*Ipvs, error) {
	log.Infof("Initialising IPVS Module...")
	if err := gipvs.Init(); err != nil {
		return nil, fmt.Errorf("IPVS initialisation failed: %v", err)
	}

	ipvs := &Ipvs{}
	if err := ipvs.Flush(); err != nil {
		return nil, fmt.Errorf("IPVS flushing table failed: %v", err)
	}

	return ipvs, nil
}

// Sync sync all ipvs rules present in state to kernel
func (ipvs *Ipvs) Sync(state state.State) error {
	ipvs.Lock()
	defer ipvs.Unlock()

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
	}

	// Cleaning rules
	for r := range rulesToRemove.Iter() {
		service := r.(types.Service)
		err := gipvs.DeleteService(*ToIpvsService(&service))
		if err != nil {
			return err
		}
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
		if err := gipvs.AddDestination(*ToIpvsService(&svc), *toIpvsDestination(&destination)); err != nil {
			return err
		}
	}

	for r := range rulesToRemove.Iter() {
		destination := r.(types.Destination)
		err := gipvs.DeleteDestination(*ToIpvsService(&svc), *toIpvsDestination(&destination))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipvs *Ipvs) addServiceAndDestinations(svc types.Service, dsts []types.Destination) error {
	ipvsService := *ToIpvsService(&svc)
	err := gipvs.AddService(ipvsService)
	if err != nil {
		return err
	}

	for _, d := range dsts {
		err := gipvs.AddDestination(ipvsService, *toIpvsDestination(&d))
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

func (ipvs *Ipvs) getCurrentServicesSet() (mapset.Set, error) {
	svcs, err := gipvs.GetServices()
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
	stateSet := mapset.NewSet()
	for _, d := range state.GetDestinations(&svc) {
		d.Name = ""
		d.ServiceId = ""
		stateSet.Add(d)
	}

	return stateSet
}

func (ipvs *Ipvs) getCurrentDestinationsSet(svc types.Service) (mapset.Set, error) {
	currentSet := mapset.NewSet()
	ipvsSvc, err := gipvs.GetService(ToIpvsService(&svc))
	if err != nil {
		return nil, err
	}

	for _, d := range ipvsSvc.Destinations {
		currentSet.Add(fromDestination(d))
	}

	return currentSet, nil
}

// Flush flushes all services and destinations from the IPVS table.
func (ipvs *Ipvs) Flush() error {
	return gipvs.Flush()
}
