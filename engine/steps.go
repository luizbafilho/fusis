package engine

import (
	"fmt"

	"github.com/luizbafilho/fusis/steps"
)

type addServiceStore struct {
	Service *Service
}

func (as addServiceStore) Do(prev steps.Result) (steps.Result, error) {
	if err := store.AddService(as.Service); err != nil {
		return nil, err
	}
	return nil, nil
}

func (as addServiceStore) Undo() error {
	fmt.Println("Deleting service from store")
	if err := store.DeleteService(as.Service); err != nil {
		return err
	}
	return nil
}

type addServiceIpvs struct {
	Service *Service
}

func (as addServiceIpvs) Do(prev steps.Result) (steps.Result, error) {
	if err := IPVSAddService(as.Service.ToIpvsService()); err != nil {
		return nil, err
	}
	return nil, nil
}

func (as addServiceIpvs) Undo() error {
	if err := store.DeleteService(as.Service); err != nil {
		return err
	}
	return nil
}
