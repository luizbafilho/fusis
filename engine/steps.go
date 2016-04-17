package engine

import (
	"fmt"

	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider"
	"github.com/luizbafilho/fusis/steps"
)

type addServiceStore struct {
	Service *ipvs.Service
}

func (as addServiceStore) Do(prev steps.Result) (steps.Result, error) {
	if err := ipvs.Store.AddService(as.Service); err != nil {
		return nil, err
	}
	return nil, nil
}

func (as addServiceStore) Undo() error {
	fmt.Println("Deleting service from ipvs.Store")
	if err := ipvs.Store.DeleteService(as.Service); err != nil {
		return err
	}
	return nil
}

type addServiceIpvs struct {
	Service *ipvs.Service
}

func (as addServiceIpvs) Do(prev steps.Result) (steps.Result, error) {
	if err := ipvs.Kernel.AddService(as.Service.ToIpvsService()); err != nil {
		return nil, err
	}
	return nil, nil
}

func (as addServiceIpvs) Undo() error {
	if err := ipvs.Store.DeleteService(as.Service); err != nil {
		return err
	}
	return nil
}

type setVip struct {
	Service *ipvs.Service
}

func (sv setVip) Do(prev steps.Result) (steps.Result, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	err = prov.AllocateVip(sv.Service)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (sv setVip) Undo() error {
	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	err = prov.ReleaseVip(*sv.Service)
	if err != nil {
		return err
	}
	return nil
}
