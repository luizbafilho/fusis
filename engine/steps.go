package engine

import (
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider"
	"github.com/luizbafilho/fusis/steps"
)

// Adding service to store
type addServiceStore struct {
	Service *ipvs.Service
}

func (as addServiceStore) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Store.AddService(as.Service)
}
func (as addServiceStore) Undo() error {
	return ipvs.Store.DeleteService(as.Service)
}

// Adding service to Ipvs module
type addServiceIpvs struct {
	Service *ipvs.Service
}

func (as addServiceIpvs) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Kernel.AddService(as.Service.ToIpvsService())
}

func (as addServiceIpvs) Undo() error {
	return ipvs.Kernel.DeleteService(as.Service.ToIpvsService())
}

//Allocating a VIP and setting it to the network interface
type setVip struct {
	Service *ipvs.Service
}

func (sv setVip) Do(prev steps.Result) (steps.Result, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	return nil, prov.AllocateVip(sv.Service)
}
func (sv setVip) Undo() error {
	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	return prov.ReleaseVip(*sv.Service)
}

// Deleting service from store
type deleteServiceStore struct {
	Service *ipvs.Service
}

func (ds deleteServiceStore) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Store.DeleteService(ds.Service)
}
func (ds deleteServiceStore) Undo() error {
	return ipvs.Store.AddService(ds.Service)
}

// Deleting service from ipvs module
type deleteServiceIpvs struct {
	Service *ipvs.Service
}

func (ds deleteServiceIpvs) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Kernel.DeleteService(ds.Service.ToIpvsService())
}
func (ds deleteServiceIpvs) Undo() error {
	return ipvs.Kernel.AddService(ds.Service.ToIpvsService())
}

//Unsetting vip
type unsetVip struct {
	Service *ipvs.Service
}

func (uv unsetVip) Do(prev steps.Result) (steps.Result, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	return nil, prov.ReleaseVip(*uv.Service)
}
func (uv unsetVip) Undo() error {
	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	return prov.AllocateVip(uv.Service)
}
