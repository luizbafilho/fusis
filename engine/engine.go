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
				fmt.Println("Stop watching changes")
				return
			}

			if err := processChange(change); err != nil {
				log.Printf("failed to process change %#v, err: %s\n", change, err)
			}
		}
	}()
}

func processChange(ch interface{}) error {
	switch change := ch.(type) {
	case store.ServiceUpsert:
		return ipvs.UpsertService(change.Service.ToIpvsService())
	}

	return nil
}
