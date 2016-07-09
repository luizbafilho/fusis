package ipvs

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	gipvs "github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api/types"
)

type Ipvs struct {
	sync.Mutex
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

type destDiffResult struct {
	toAdd    []*types.Destination
	toRemove []*types.Destination
	toUpdate []*types.Destination
}

func (ipvs *Ipvs) diffDestinations(old, new *types.Service) destDiffResult {
	oldDests := old.Destinations
	newDests := new.Destinations
	toAddMap := make(map[string]*types.Destination)
	for i, d := range newDests {
		toAddMap[d.KernelKey()] = &newDests[i]
	}
	var toAdd, toRemove, toUpdate []*types.Destination
	for i, d := range oldDests {
		key := d.KernelKey()
		if newDest, isPresent := toAddMap[key]; isPresent {
			toUpdate = append(toUpdate, newDest)
			delete(toAddMap, key)
		} else {
			toRemove = append(toRemove, &oldDests[i])
		}
	}
	for _, d := range toAddMap {
		toAdd = append(toAdd, d)
	}
	return destDiffResult{
		toAdd:    toAdd,
		toRemove: toRemove,
		toUpdate: toUpdate,
	}
}

func (ipvs *Ipvs) SyncState(state State) error {
	oldServices, err := gipvs.GetServices()
	if err != nil {
		return err
	}
	newServices := state.GetServices()
	toAddMap := make(map[string]*types.Service)
	for i, s := range newServices {
		toAddMap[s.KernelKey()] = &newServices[i]
	}
	var toAdd, toRemove []*types.Service
	var toMerge [][]*types.Service
	for _, gipvsSvc := range oldServices {
		s := FromService(gipvsSvc)
		key := s.KernelKey()
		if newService, isPresent := toAddMap[key]; isPresent {
			toMerge = append(toMerge, []*types.Service{&s, newService})
			delete(toAddMap, key)
		} else {
			toRemove = append(toRemove, &s)
		}
	}
	for _, s := range toAddMap {
		toAdd = append(toAdd, s)
	}
	var errors []string
	for _, s := range toAdd {
		err = gipvs.AddService(*ToIpvsService(s))
		if err != nil {
			errors = append(errors, fmt.Sprintf("error adding service %#v: %s", s, err))
		}
	}
	for _, s := range toRemove {
		err = gipvs.DeleteService(*ToIpvsService(s))
		if err != nil {
			errors = append(errors, fmt.Sprintf("error deleting service %#v: %s", s, err))
		}
	}
	for _, services := range toMerge {
		oldService := services[0]
		newService := services[1]
		newGipvsService := *ToIpvsService(newService)
		err = gipvs.UpdateService(newGipvsService)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error updating service %#v: %s", newService, err))
		}
		result := ipvs.diffDestinations(oldService, newService)
		for _, d := range result.toAdd {
			err = gipvs.AddDestination(newGipvsService, *toIpvsDestination(d))
			if err != nil {
				errors = append(errors, fmt.Sprintf("error adding destination %#v: %s", d, err))
			}
		}
		for _, d := range result.toRemove {
			err = gipvs.DeleteDestination(newGipvsService, *toIpvsDestination(d))
			if err != nil {
				errors = append(errors, fmt.Sprintf("error deleting destination %#v: %s", d, err))
			}
		}
		for _, d := range result.toUpdate {
			err = gipvs.UpdateDestination(newGipvsService, *toIpvsDestination(d))
			if err != nil {
				errors = append(errors, fmt.Sprintf("error deleting destination %#v: %s", d, err))
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("multiple errors: %s", strings.Join(errors, " | "))
	}
	return nil
}

// Flush flushes all services and destinations from the IPVS table.
func (ipvs *Ipvs) Flush() error {
	return gipvs.Flush()
}
