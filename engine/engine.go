package engine

import (
	"fmt"
	"log"

	"github.com/luizbafilho/janus/ipvs"
	"github.com/luizbafilho/janus/store"
)

type EngineService struct {
	store          store.Store
	changesChannel chan interface{}
}

func NewEngine(store store.Store) EngineService {
	return EngineService{
		store:          store,
		changesChannel: make(chan interface{}),
	}
}

func (es EngineService) Serve() {
	es.subscribeStore()
	es.handleChanges()
}

func (es EngineService) subscribeStore() {
	go func() {
		err := es.store.Subscribe(es.changesChannel)
		if err != nil {
			log.Println("Error on Subscribe to store")
		}
	}()
}

func (es EngineService) handleChanges() {
	go func() {
		for {
			change := <-es.changesChannel

			if change == nil {
				fmt.Println("Change was nil")
				continue
			}

			if err := processChange(change); err != nil {
				log.Printf("failed to process change %#v, err: %s\n", change, err)
			}
		}
	}()
}

func processChange(ch interface{}) error {
	switch change := ch.(type) {
	case store.ServiceEvent:
		processServiceEvent(change)
	case store.DestinationEvent:
		processDestinationEvent(change)
	}

	return nil
}

func processServiceEvent(se store.ServiceEvent) error {
	switch se.Action {
	case store.CreateEvent:
		return ipvs.AddService(se.Service.ToIpvsService())
	case store.UpdateEvent:
		return ipvs.UpdateService(se.Service.ToIpvsService())
	case store.DeleteEvent:
		return ipvs.DeleteService(se.Service.ToIpvsService())
	}
	return nil
}

func processDestinationEvent(de store.DestinationEvent) error {
	switch de.Action {
	case store.CreateEvent:
		return ipvs.AddDestination(de.Service.ToIpvsService(), *de.Destination.ToIpvsDestination())
	case store.UpdateEvent:
		return ipvs.UpdateDestination(de.Service.ToIpvsService(), *de.Destination.ToIpvsDestination())
	case store.DeleteEvent:
		return ipvs.DeleteDestination(de.Service.ToIpvsService(), *de.Destination.ToIpvsDestination())
	}
	return nil
}
